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

	// For remote files, download and stream
	resp, err := http.Get(linkInfo.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to download file")
	}

	// Copy response body to writer
	_, err = io.Copy(writer, resp.Body)
	return err
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", linkInfo.Url, nil)
	if err != nil {
		return err
	}

	// Set range header
	rangeHeader := "bytes=" + fmt.Sprintf("%d-%d", offset, offset+length-1)
	req.Header.Set("Range", rangeHeader)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return errors.New("failed to download file range")
	}

	// Copy response body to writer
	_, err = io.Copy(writer, resp.Body)
	return err
}