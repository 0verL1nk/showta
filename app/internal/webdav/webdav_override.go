package webdav

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"overlink.top/app/system/logic"
	"overlink.top/app/system/msg"
)

func (h *Handler) ServeHTTPOverride(w http.ResponseWriter, r *http.Request) {
	status, err := http.StatusBadRequest, errUnsupportedMethod
	if h.FileSystem == nil {
		status, err = http.StatusInternalServerError, errNoFileSystem
	} else if h.LockSystem == nil {
		status, err = http.StatusInternalServerError, errNoLockSystem
	} else {
		switch r.Method {
		case "OPTIONS":
			status, err = h.handleOptions(w, r)
		case "GET", "HEAD", "POST":
			status, err = h.handleGetHeadPostOverride(w, r)
		case "DELETE":
			status, err = h.handleDeleteOverride(w, r)
		case "PUT":
			status, err = h.handlePutOverride(w, r)
		case "MKCOL":
			status, err = h.handleMkcolOverride(w, r)
		case "COPY", "MOVE":
			status, err = h.handleCopyMoveOverride(w, r)
		case "LOCK":
			status, err = h.handleLock(w, r)
		case "UNLOCK":
			status, err = h.handleUnlock(w, r)
		case "PROPFIND":
			status, err = h.handlePropfindOverride(w, r)
		case "PROPPATCH":
			status, err = h.handleProppatch(w, r)
		}
	}

	if status != 0 {
		w.WriteHeader(status)
		if status != http.StatusNoContent {
			w.Write([]byte(StatusText(status)))
		}
	}
	if h.Logger != nil {
		h.Logger(r, err)
	}
}

func (h *Handler) handleGetHeadPostOverride(w http.ResponseWriter, r *http.Request) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}
	// TODO: check locks for read-only access??
	ctx := r.Context()

	// Try to get file info from cache first
	fi, err := GetCachedFileInfo(ctx, reqPath)
	if err != nil || fi == nil {
		// Not in cache, fetch it
		fi, err = logic.GetFile(ctx, reqPath)
		if err != nil {
			return http.StatusNotFound, err
		}
		// Cache the file info for future use
		CacheFileInfo(reqPath, fi)
	}

	if fi.IsDir() {
		return http.StatusMethodNotAllowed, nil
	}
	etag, err := findETagOverride(ctx, h.FileSystem, h.LockSystem, reqPath, fi)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	w.Header().Set("ETag", etag)
	// Let ServeContent determine the Content-Type header.
	// http.ServeContent(w, r, reqPath, fi.ModTime(), f)
	logic.ProxyFile(r, w, reqPath)

	return 0, nil
}

// handleDeleteOverride handles DELETE requests with cache invalidation
func (h *Handler) handleDeleteOverride(w http.ResponseWriter, r *http.Request) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}
	release, status, err := h.confirmLocks(r, reqPath, "")
	if err != nil {
		return status, err
	}
	defer release()

	ctx := r.Context()

	// TODO: return MultiStatus where appropriate.

	// "godoc os RemoveAll" says that "If the path does not exist, RemoveAll
	// returns nil (no error)." WebDAV semantics are that it should return a
	// "404 Not Found". We therefore have to Stat before we RemoveAll.
	if _, err := h.FileSystem.Stat(ctx, reqPath); err != nil {
		if os.IsNotExist(err) {
			return http.StatusNotFound, err
		}
		return http.StatusMethodNotAllowed, err
	}

	// Invalidate cache for the deleted path and its parent directory
	InvalidateCache(reqPath)
	parentDir := path.Dir(reqPath)
	InvalidateCache(parentDir)

	if err := h.FileSystem.RemoveAll(ctx, reqPath); err != nil {
		return http.StatusMethodNotAllowed, err
	}
	return http.StatusNoContent, nil
}

// handlePutOverride handles PUT requests with cache invalidation
func (h *Handler) handlePutOverride(w http.ResponseWriter, r *http.Request) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}
	release, status, err := h.confirmLocks(r, reqPath, "")
	if err != nil {
		return status, err
	}
	defer release()
	// TODO(rost): Support the If-Match, If-None-Match headers? See bradfitz'
	// comments in http.checkEtag.
	ctx := r.Context()

	f, err := h.FileSystem.OpenFile(ctx, reqPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return http.StatusNotFound, err
	}
	_, copyErr := io.Copy(f, r.Body)
	fi, statErr := f.Stat()
	closeErr := f.Close()

	// Invalidate cache for the modified path and its parent directory
	InvalidateCache(reqPath)
	parentDir := path.Dir(reqPath)
	InvalidateCache(parentDir)

	// TODO(rost): Returning 405 Method Not Allowed might not be appropriate.
	if copyErr != nil {
		return http.StatusMethodNotAllowed, copyErr
	}
	if statErr != nil {
		return http.StatusMethodNotAllowed, statErr
	}
	if closeErr != nil {
		return http.StatusMethodNotAllowed, closeErr
	}
	etag, err := findETag(ctx, h.FileSystem, h.LockSystem, reqPath, fi)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	w.Header().Set("ETag", etag)
	return http.StatusCreated, nil
}

// handleMkcolOverride handles MKCOL requests with cache invalidation
func (h *Handler) handleMkcolOverride(w http.ResponseWriter, r *http.Request) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}
	release, status, err := h.confirmLocks(r, reqPath, "")
	if err != nil {
		return status, err
	}
	defer release()

	ctx := r.Context()

	if r.ContentLength > 0 {
		return http.StatusUnsupportedMediaType, nil
	}

	// Invalidate cache for the parent directory
	parentDir := path.Dir(reqPath)
	InvalidateCache(parentDir)

	if err := h.FileSystem.Mkdir(ctx, reqPath, 0777); err != nil {
		if os.IsNotExist(err) {
			return http.StatusConflict, err
		}
		return http.StatusMethodNotAllowed, err
	}
	return http.StatusCreated, nil
}

// handleCopyMoveOverride handles COPY and MOVE requests with cache invalidation
func (h *Handler) handleCopyMoveOverride(w http.ResponseWriter, r *http.Request) (status int, err error) {
	hdr := r.Header.Get("Destination")
	if hdr == "" {
		return http.StatusBadRequest, errInvalidDestination
	}
	u, err := url.Parse(hdr)
	if err != nil {
		return http.StatusBadRequest, errInvalidDestination
	}
	if u.Host != "" && u.Host != r.Host {
		return http.StatusBadGateway, errInvalidDestination
	}

	src, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}

	dst, status, err := h.stripPrefix(u.Path)
	if err != nil {
		return status, err
	}

	if dst == "" {
		return http.StatusBadGateway, errInvalidDestination
	}
	if dst == src {
		return http.StatusForbidden, errDestinationEqualsSource
	}

	ctx := r.Context()

	if r.Method == "COPY" {
		// Section 7.5.1 says that a COPY only needs to lock the destination,
		// not both destination and source. Strictly speaking, this is racy,
		// even though a COPY doesn't modify the source, if a concurrent
		// operation modifies the source. However, the litmus test explicitly
		// checks that COPYing a locked-by-another source is OK.
		release, status, err := h.confirmLocks(r, "", dst)
		if err != nil {
			return status, err
		}
		defer release()

		// Section 9.8.3 says that "The COPY method on a collection without a Depth
		// header must act as if a Depth header with value "infinity" was included".
		depth := infiniteDepth
		if hdr := r.Header.Get("Depth"); hdr != "" {
			depth = parseDepth(hdr)
			if depth != 0 && depth != infiniteDepth {
				// Section 9.8.3 says that "A client may submit a Depth header on a
				// COPY on a collection with a value of "0" or "infinity"."
				return http.StatusBadRequest, errInvalidDepth
			}
		}

		// Invalidate cache for destination and its parent directory
		InvalidateCache(dst)
		dstParentDir := path.Dir(dst)
		InvalidateCache(dstParentDir)

		return copyFiles(ctx, h.FileSystem, src, dst, r.Header.Get("Overwrite") != "F", depth, 0)
	}

	release, status, err := h.confirmLocks(r, src, dst)
	if err != nil {
		return status, err
	}
	defer release()

	// Section 9.9.2 says that "The MOVE method on a collection must act as if
	// a "Depth: infinity" header was used on it. A client must not submit a
	// Depth header on a MOVE on a collection with any value but "infinity"."
	if hdr := r.Header.Get("Depth"); hdr != "" {
		if parseDepth(hdr) != infiniteDepth {
			return http.StatusBadRequest, errInvalidDepth
		}
	}

	// Invalidate cache for source, destination, and their parent directories
	InvalidateCache(src)
	srcParentDir := path.Dir(src)
	InvalidateCache(srcParentDir)
	InvalidateCache(dst)
	dstParentDir := path.Dir(dst)
	InvalidateCache(dstParentDir)

	return moveFiles(ctx, h.FileSystem, src, dst, r.Header.Get("Overwrite") == "T")
}

func (h *Handler) handlePropfindOverride(w http.ResponseWriter, r *http.Request) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}
	ctx := r.Context()

	// Try to get file info from cache first
	fi, err := GetCachedFileInfo(ctx, reqPath)
	if err != nil || fi == nil {
		// Not in cache, fetch it
		fi, err = logic.GetFile(ctx, reqPath)
		if err != nil {
			// if os.IsNotExist(err) {
			return http.StatusNotFound, err
			// }
			// return http.StatusMethodNotAllowed, err
		}
		// Cache the file info for future use
		CacheFileInfo(reqPath, fi)
	}

	depth := infiniteDepth
	if hdr := r.Header.Get("Depth"); hdr != "" {
		depth = parseDepth(hdr)
		if depth == invalidDepth {
			return http.StatusBadRequest, errInvalidDepth
		}
	}
	pf, status, err := readPropfind(r.Body)
	if err != nil {
		return status, err
	}

	mw := multistatusWriter{w: w}

	walkFn := func(reqPath string, info msg.Finfo, err error) error {
		if err != nil {
			return err
			// return handlePropfindError(err, info)
		}

		// Cache the file info for this path as well
		CacheFileInfo(reqPath, info)

		var pstats []Propstat
		if pf.Propname != nil {
			pnames, err := propnames(ctx, h.FileSystem, h.LockSystem, reqPath)
			if err != nil {
				return err
				// return handlePropfindError(err, info)
			}
			pstat := Propstat{Status: http.StatusOK}
			for _, xmlname := range pnames {
				pstat.Props = append(pstat.Props, Property{XMLName: xmlname})
			}
			pstats = append(pstats, pstat)
		} else if pf.Allprop != nil {
			pstats, err = allpropOverride(ctx, info, h.LockSystem, reqPath, pf.Prop)
		} else {
			pstats, err = propsOverride(ctx, info, h.LockSystem, reqPath, pf.Prop)
		}
		if err != nil {
			return err
			// return handlePropfindError(err, info)
		}
		href := path.Join(h.Prefix, reqPath)
		if href != "/" && info.IsDir() {
			href += "/"
		}
		return mw.write(makePropstatResponse(href, pstats))
	}

	walkErr := walkFSOverride(ctx, h.FileSystem, depth, reqPath, fi, walkFn)
	closeErr := mw.close()
	if walkErr != nil {
		return http.StatusInternalServerError, walkErr
	}
	if closeErr != nil {
		return http.StatusInternalServerError, closeErr
	}
	return 0, nil
}
