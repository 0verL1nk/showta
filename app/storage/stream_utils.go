package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// DefaultStreamFile provides a default implementation for StreamFile
// that uses the Link method and streams from the URL
func DefaultStreamFile(ctx context.Context, store Storage, rpath string, writer io.Writer) error {
	// Get file info
	getter, ok := store.(Getter)
	if !ok {
		return errors.New("storage does not implement Getter interface")
	}

	info, err := getter.Get(rpath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("cannot stream a directory")
	}

	// Get link to file
	linkInfo, err := store.Link(info)
	if err != nil {
		return err
	}

	// If it's a direct file (local storage), stream directly
	if store.IsDirect() {
		// For direct files, we can open and stream directly
		// This is a simplified version - in practice, this would depend on the specific implementation
		return errors.New("direct file streaming not implemented in default handler")
	}

	// For remote files, download and stream with context awareness
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, linkInfo.Url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to download file")
	}

	return copyWithContext(ctx, writer, resp.Body)
}

// DefaultStreamRange provides a default implementation for StreamRange
// that uses the Link method and streams a range from the URL
func DefaultStreamRange(ctx context.Context, store Storage, rpath string, offset, length int64, writer io.Writer) error {
	// Get file info
	getter, ok := store.(Getter)
	if !ok {
		return errors.New("storage does not implement Getter interface")
	}

	info, err := getter.Get(rpath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("cannot stream a directory")
	}

	// Get link to file
	linkInfo, err := store.Link(info)
	if err != nil {
		return err
	}

	// If it's a direct file (local storage), stream directly
	if store.IsDirect() {
		// For direct files, we can open and stream directly
		// This is a simplified version - in practice, this would depend on the specific implementation
		return errors.New("direct file range streaming not implemented in default handler")
	}

	// For remote files, download range and stream
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, linkInfo.Url, nil)
	if err != nil {
		return err
	}

	// Set range header
	rangeHeader := "bytes=" + fmt.Sprintf("%d-%d", offset, offset+length-1)
	req.Header.Set("Range", rangeHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return errors.New("failed to download file range")
	}

	return copyWithContext(ctx, writer, resp.Body)
}

// copyWithContext copies data from reader to writer while monitoring context cancellation
func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) error {
	const bufferSize = 64 * 1024 // 64KB buffer
	buf := make([]byte, bufferSize)

	for {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}

		nr, er := src.Read(buf)
		if nr > 0 {
			if _, ew := dst.Write(buf[:nr]); ew != nil {
				return ew
			}
		}

		if er != nil {
			if errors.Is(er, io.EOF) {
				break
			}
			return er
		}
	}

	return nil
}
