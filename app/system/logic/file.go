package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"overlink.top/app/internal/memcache"
	"overlink.top/app/internal/sign"
	"overlink.top/app/lib/util"
	"overlink.top/app/storage"
	"overlink.top/app/system/conf"
	"overlink.top/app/system/log"
	"overlink.top/app/system/model"
	"overlink.top/app/system/msg"
	"path"
	"strconv"
	"strings"
	"time"
)

func ListFile(ctx context.Context, rpath string) (list []msg.Finfo, err error) {
	rpath = util.StandardPath(rpath)
	//Virtual mounting directory
	if rpath == "/" {
		storageMap.Range(func(key, value interface{}) bool {
			v := value.(storage.Storage)
			mountPath := v.GetData().MountPath
			list = append(list, &msg.FileInfo{
				Path:     mountPath,
				Name:     util.SimplePath(mountPath),
				Size:     0,
				Modified: v.GetData().UpdatedAt,
				IsFolder: true,
			})

			return true
		})
	} else {
		var store storage.Storage
		storageMap.Range(func(key, value interface{}) bool {
			v := value.(storage.Storage)
			mountPath := v.GetData().MountPath
			if strings.Contains(rpath, mountPath) {
				store = v
				return false
			}

			return true
		})

		if store != nil {
			if store.AllowCache() {
				list, err = cacheListFile(rpath, store)
			} else {
				list, err = store.List(&msg.FileInfo{Path: rpath})
			}
		}
	}

	return
}

func ViewListFile(ctx context.Context, rpath string) (resp msg.ListFileResp, err error) {
	list, err := ListFile(ctx, rpath)
	if err != nil {
		return
	}

	setting, err := model.GetFolderSettingByFolder(rpath)
	if err != nil {
		return
	}

	var dataList []msg.FileInfo
	for _, v := range list {
		dataList = append(dataList, msg.FileInfo{
			Path:     v.GetPath(),
			Name:     v.GetName(),
			Size:     v.GetSize(),
			Modified: v.ModTime(),
			IsFolder: v.IsDir(),
		})
	}
	resp.List = dataList
	if setting.ID > 0 {
		resp.Topmd = setting.Topmd
		resp.Readme = setting.Readme
	}

	return
}

func GetStorageFile(c *gin.Context, rpath string) (resp msg.GetFileResp, err error) {
	info, err := GetFile(c, rpath)
	if err != nil {
		return
	}

	isDir := info.IsDir()
	name := info.GetName()
	rawUrl := info.GetRaw()
	var ptype int
	if !isDir {
		ptype = getPreviewType(name)
		if rawUrl == "" {
			var param string
			if conf.GlobalSign {
				param = "?sig=" + sign.Gen(rpath, "")
			}

			rawUrl = fmt.Sprintf("%s/fd%s%s", getHost(c.Request), rpath, param)
		}
	}

	resp = msg.GetFileResp{
		FileInfo: msg.FileInfo{
			Path:     info.GetPath(),
			Name:     name,
			Size:     info.GetSize(),
			Modified: info.ModTime(),
			IsFolder: isDir,
			Ptype:    ptype,
		},
		RawUrl: rawUrl,
	}

	return
}

func ProxyFile(r *http.Request, w http.ResponseWriter, rpath string) {
	rpath = util.StandardPath(rpath)
	var store storage.Storage
	storageMap.Range(func(key, value interface{}) bool {
		v := value.(storage.Storage)
		mountPath := v.GetData().MountPath
		if strings.Contains(rpath, mountPath) {
			store = v
			return false
		}

		return true
	})

	if store == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "no such file:", rpath)
		return
	}

	// Try to use streaming if available
	if streamer, ok := store.(interface {
		StreamFile(ctx context.Context, path string, writer io.Writer) error
	}); ok {
		info, err := GetFile(r.Context(), rpath)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "file not found:", err)
			return
		}

		if info.IsDir() {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "cannot stream a directory")
			return
		}

		fileInfo := &fileInfoAdapter{info}
		totalSize := info.GetSize()
		rangeHeader := r.Header.Get("Range")

		if rangeHeader != "" {
			if ranger, ok := store.(interface {
				StreamRange(ctx context.Context, path string, offset, length int64, writer io.Writer) error
			}); ok {
				byteRange, err := parseRangeHeader(rangeHeader, totalSize)
				if err != nil {
					w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", totalSize))
					w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
					return
				}

				setAttach(w, fileInfo)
				length := byteRange.Length()
				w.Header().Set("Accept-Ranges", "bytes")
				w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", byteRange.Start, byteRange.End, totalSize))
				w.WriteHeader(http.StatusPartialContent)

				err = ranger.StreamRange(r.Context(), rpath, byteRange.Start, length, w)
				if err != nil {
					if isClientDisconnectError(err) {
						log.Debugf("client disconnected during range streaming: %s", rpath)
					} else {
						log.Errorf("range streaming file error: %v", err)
					}
				}
				return
			}
		}

		// Normal full-file streaming path
		setAttach(w, fileInfo)
		w.Header().Set("Accept-Ranges", "bytes")

		err = streamer.StreamFile(r.Context(), rpath, w)
		if err != nil {
			if isClientDisconnectError(err) {
				log.Debugf("client disconnected during streaming: %s", rpath)
			} else {
				log.Errorf("streaming file error: %v", err)
			}
			return
		}
		return
	}
	log.Debugf("streaming not supported for storage backend, falling back to original method")
	// Fall back to original method
	proxyFileOriginal(r, w, rpath, store)
}

// proxyFileOriginal is the original implementation for backward compatibility
func proxyFileOriginal(r *http.Request, w http.ResponseWriter, rpath string, store storage.Storage) {
	var linkInfo *msg.LinkInfo
	var err error
	if store.AllowCache() {
		item, err := findCacheFile(rpath, store)
		if err != nil || item.IsDir() {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "link file error:", err)
			return
		}

		linkInfo, err = cacheFileLink(item, store)
	} else {
		linkInfo, err = store.Link(&msg.FileInfo{Path: rpath})
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "link file error:", err)
		return
	}

	link := linkInfo.Url
	if !store.IsDirect() {
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	f, err := os.Open(link)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "open file error:", err)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "call stat error:", err)
		return
	}

	if fi.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "no such file:", link)
		return
	}

	setAttach(w, fi)

	// Use buffered copy for better memory usage with chunked transfer encoding
	bufferedCopyWithChunkedEncoding(w, f, fi.Size())
}

// bufferedCopy copies data from reader to writer using a fixed-size buffer
// to optimize memory usage for large file transfers
func bufferedCopy(w io.Writer, r io.Reader) error {
	// Use configured buffer size or default to 64KB
	bufferSize := 64 * 1024
	if conf.AppConf.WebDAV.BufferSize > 0 {
		bufferSize = conf.AppConf.WebDAV.BufferSize
	}
	buf := make([]byte, bufferSize)
	_, err := io.CopyBuffer(w, r, buf)
	return err
}

// bufferedCopyWithChunkedEncoding copies data from reader to writer using a fixed-size buffer
// and supports chunked transfer encoding for large files
func bufferedCopyWithChunkedEncoding(w http.ResponseWriter, r io.Reader, fileSize int64) error {
	// For large files, we rely on Go's automatic chunked transfer encoding
	// when Content-Length is not explicitly set or when using HTTP/1.1+

	// Use configured buffer size or default to 64KB
	bufferSize := 64 * 1024
	if conf.AppConf.WebDAV.BufferSize > 0 {
		bufferSize = conf.AppConf.WebDAV.BufferSize
	}
	buf := make([]byte, bufferSize)
	_, err := io.CopyBuffer(w, r, buf)
	return err
}

// fileInfoAdapter adapts msg.Finfo to os.FileInfo interface
type fileInfoAdapter struct {
	info msg.Finfo
}

func (f *fileInfoAdapter) Name() string       { return f.info.GetName() }
func (f *fileInfoAdapter) Size() int64        { return f.info.GetSize() }
func (f *fileInfoAdapter) Mode() os.FileMode  { return 0644 }
func (f *fileInfoAdapter) ModTime() time.Time { return f.info.ModTime() }
func (f *fileInfoAdapter) IsDir() bool        { return f.info.IsDir() }
func (f *fileInfoAdapter) Sys() interface{}   { return nil }

func setAttach(w http.ResponseWriter, fi os.FileInfo) {
	name := fi.Name()
	ext := path.Ext(name)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, name, url.QueryEscape(name)))
	w.Header().Set("Content-Type", mimeType)
}

func GetFile(ctx context.Context, rpath string) (info msg.Finfo, err error) {
	rpath = util.StandardPath(rpath)
	//Virtual mounting directory
	if rpath == "/" {
		info = &msg.FileInfo{
			Name:     "/",
			IsFolder: true,
		}
		return
	}

	var store storage.Storage
	storageMap.Range(func(key, value interface{}) bool {
		v := value.(storage.Storage)
		mountPath := v.GetData().MountPath
		if strings.Contains(rpath, mountPath) {
			store = v
			return false
		}

		return true
	})

	if store == nil {
		err = errors.New("dir not exist")
		return
	}

	getter, ok := store.(storage.Getter)
	if ok {
		info, err = getter.Get(rpath)
		return
	}

	if store.AllowCache() {
		item, err := findCacheFile(rpath, store)
		if err != nil {
			return nil, err
		}

		if item.IsDir() {
			return item, nil
		}

		linkInfo, err := cacheFileLink(item, store)
		if err != nil {
			return nil, err
		}

		info = &msg.FileInfo{
			FileId:   item.GetFileId(),
			Path:     item.GetPath(),
			Name:     item.GetName(),
			Size:     item.GetSize(),
			Modified: item.ModTime(),
			IsFolder: item.IsDir(),
			RawUrl:   linkInfo.Url,
		}

		return info, nil
	}

	err = errors.New("dir err")
	return
}

func Subdir(ctx context.Context, rpath string) (list []msg.Finfo, err error) {
	fileList, err := ListFile(ctx, rpath)
	if err != nil {
		return
	}

	for _, v := range fileList {
		if v.IsDir() {
			list = append(list, v)
		}
	}

	return
}

func cacheListFile(rpath string, store storage.Storage) (list []msg.Finfo, err error) {
	data, found := memcache.Get(memcache.List, rpath)
	if found {
		log.Debugf("[CACHE] list path: %+s", rpath)
		list = data.([]msg.Finfo)
		return
	}

	info := msg.FileInfo{Path: rpath}
	parentPath := util.GetParentDir(rpath)
	if parentPath != "/" {
		pdata, pfound := memcache.Get(memcache.List, parentPath)
		if pfound {
			plist := pdata.([]msg.Finfo)
			for _, item := range plist {
				if item.GetPath() == rpath {
					info.FileId = item.GetFileId()
					break
				}
			}
		}
	}

	list, err = store.List(&info)
	if err != nil {
		return
	}

	memcache.Set(memcache.List, rpath, list)
	return
}

func findCacheFile(rpath string, store storage.Storage) (info msg.Finfo, err error) {
	dpath, fname := util.SplitPath(rpath)
	if dpath == "/" {
		err = errors.New("dir err")
		return
	}

	list, err := cacheListFile(dpath, store)
	if err != nil {
		return nil, err
	}

	for _, item := range list {
		if item.GetName() == fname {
			return item, nil
		}
	}

	return
}

func cacheFileLink(info msg.Finfo, store storage.Storage) (linkInfo *msg.LinkInfo, err error) {
	rpath := info.GetPath()
	data, found := memcache.Get(memcache.Link, rpath)
	if found {
		log.Debugf("[CACHE] link path: %+s", rpath)
		linkInfo = data.(*msg.LinkInfo)
		return
	}

	linkInfo, err = store.Link(info)
	if err != nil {
		return nil, err
	}

	memcache.Expire(memcache.Link, rpath, linkInfo, linkInfo.Expire)
	return
}

// isClientDisconnectError checks if an error indicates the client closed the connection
func isClientDisconnectError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	patterns := []string{
		"broken pipe",
		"connection reset by peer",
		"wsasend",
		"forcibly closed",
		"use of closed network connection",
		"connection refused",
		"EOF",
	}

	for _, pattern := range patterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

type byteRange struct {
	Start int64
	End   int64
}

func (br byteRange) Length() int64 {
	return br.End - br.Start + 1
}

func parseRangeHeader(header string, size int64) (byteRange, error) {
	const prefix = "bytes="
	if !strings.HasPrefix(header, prefix) {
		return byteRange{}, errors.New("invalid range unit")
	}

	spec := strings.TrimSpace(header[len(prefix):])
	if spec == "" {
		return byteRange{}, errors.New("empty range")
	}

	if idx := strings.Index(spec, ","); idx >= 0 {
		spec = spec[:idx]
	}

	var start, end int64
	if strings.HasPrefix(spec, "-") {
		length, err := strconv.ParseInt(strings.TrimPrefix(spec, "-"), 10, 64)
		if err != nil || length <= 0 {
			return byteRange{}, errors.New("invalid suffix range")
		}
		if length > size {
			length = size
		}
		start = size - length
		end = size - 1
	} else if strings.HasSuffix(spec, "-") {
		value := strings.TrimSuffix(spec, "-")
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed < 0 {
			return byteRange{}, errors.New("invalid range start")
		}
		if parsed >= size {
			return byteRange{}, errors.New("range start beyond size")
		}
		start = parsed
		end = size - 1
	} else {
		parts := strings.Split(spec, "-")
		if len(parts) != 2 {
			return byteRange{}, errors.New("invalid range format")
		}
		var err error
		start, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil || start < 0 {
			return byteRange{}, errors.New("invalid range start")
		}
		end, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil || end < start {
			return byteRange{}, errors.New("invalid range end")
		}
		if start >= size {
			return byteRange{}, errors.New("range start beyond size")
		}
		if end >= size {
			end = size - 1
		}
	}

	return byteRange{Start: start, End: end}, nil
}
