package webdav

import (
	"context"
	"path"
	"path/filepath"
	"overlink.top/app/system/logic"
	"overlink.top/app/system/msg"
	"sort"
)

type WalkFunc func(pathStr string, info msg.Finfo, err error) error

// Pagination parameters for directory listing
const (
	DefaultPageSize = 1000
	MaxPageSize     = 10000
)

// PagedDirectoryListing represents a paginated directory listing
type PagedDirectoryListing struct {
	Items      []msg.Finfo
	TotalCount int
	Page       int
	PageSize   int
	HasMore    bool
}

// walkFSOverride walks the file system with the given depth
func walkFSOverride(ctx context.Context, fs FileSystem, depth int, name string, info msg.Finfo, walkFn WalkFunc) error {
	// This implementation is based on Walk's code in the standard path/filepath package.
	err := walkFn(name, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !info.IsDir() || depth == 0 {
		return nil
	}
	if depth == 1 {
		depth = 0
	}

	// Read directory names.
	// Try to get directory listing from cache first
	fileInfos, err := GetCachedDirList(ctx, name)
	if err != nil || fileInfos == nil {
		// Not in cache, fetch it
		fileInfos, err = logic.ListFile(ctx, name)
		if err != nil {
			return walkFn(name, info, err)
		}
		// Cache the directory listing for future use
		CacheDirList(name, fileInfos)
	}

	// Sort file infos by name for consistent ordering
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].GetName() < fileInfos[j].GetName()
	})

	for _, fileInfo := range fileInfos {
		filename := path.Join(name, fileInfo.GetName())
		// fileInfo, err := fs.Stat(ctx, filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walkFSOverride(ctx, fs, depth, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

// getPagedDirectoryListing gets a paginated directory listing
func getPagedDirectoryListing(ctx context.Context, name string, page, pageSize int) (*PagedDirectoryListing, error) {
	// Validate page parameters
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	// Try to get directory listing from cache first
	fileInfos, err := GetCachedDirList(ctx, name)
	if err != nil || fileInfos == nil {
		// Not in cache, fetch it
		fileInfos, err = logic.ListFile(ctx, name)
		if err != nil {
			return nil, err
		}
		// Cache the directory listing for future use
		CacheDirList(name, fileInfos)
	}

	// Sort file infos by name for consistent ordering
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].GetName() < fileInfos[j].GetName()
	})

	// Calculate pagination
	totalCount := len(fileInfos)
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= totalCount {
		// Return empty page
		return &PagedDirectoryListing{
			Items:      []msg.Finfo{},
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
			HasMore:    false,
		}, nil
	}

	if endIndex > totalCount {
		endIndex = totalCount
	}

	hasMore := endIndex < totalCount
	items := fileInfos[startIndex:endIndex]

	return &PagedDirectoryListing{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		HasMore:    hasMore,
	}, nil
}
