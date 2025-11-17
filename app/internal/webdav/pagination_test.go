package webdav

import (
	"context"
	"testing"

	"overlink.top/app/system/msg"
)

func TestGetPagedDirectoryListing(t *testing.T) {
	// Create mock directory listing with multiple files
	mockFiles := make([]msg.Finfo, 150)
	for i := 0; i < 150; i++ {
		mockFiles[i] = &mockFileInfo{
			name:    "file" + string(rune(i+'0')) + ".txt",
			size:    int64(i * 100),
			isDir:   i%10 == 0, // Every 10th item is a directory
			path:    "/test/large_directory/file" + string(rune(i+'0')) + ".txt",
		}
	}

	// Cache the directory listing
	dirPath := "/test/large_directory"
	CacheDirList(dirPath, mockFiles)

	// Test first page
	page1, err := getPagedDirectoryListing(context.Background(), dirPath, 1, 50)
	if err != nil {
		t.Errorf("getPagedDirectoryListing failed for page 1: %v", err)
	}

	if page1.TotalCount != 150 {
		t.Errorf("Expected total count 150, got %d", page1.TotalCount)
	}

	if len(page1.Items) != 50 {
		t.Errorf("Expected 50 items on page 1, got %d", len(page1.Items))
	}

	if !page1.HasMore {
		t.Error("Expected HasMore to be true for page 1")
	}

	if page1.Page != 1 {
		t.Errorf("Expected page 1, got %d", page1.Page)
	}

	if page1.PageSize != 50 {
		t.Errorf("Expected page size 50, got %d", page1.PageSize)
	}

	// Test second page
	page2, err := getPagedDirectoryListing(context.Background(), dirPath, 2, 50)
	if err != nil {
		t.Errorf("getPagedDirectoryListing failed for page 2: %v", err)
	}

	if len(page2.Items) != 50 {
		t.Errorf("Expected 50 items on page 2, got %d", len(page2.Items))
	}

	if !page2.HasMore {
		t.Error("Expected HasMore to be true for page 2")
	}

	// Test third page
	page3, err := getPagedDirectoryListing(context.Background(), dirPath, 3, 50)
	if err != nil {
		t.Errorf("getPagedDirectoryListing failed for page 3: %v", err)
	}

	if len(page3.Items) != 50 {
		t.Errorf("Expected 50 items on page 3, got %d", len(page3.Items))
	}

	if page3.HasMore {
		t.Error("Expected HasMore to be false for page 3")
	}

	// Test fourth page (should be empty)
	page4, err := getPagedDirectoryListing(context.Background(), dirPath, 4, 50)
	if err != nil {
		t.Errorf("getPagedDirectoryListing failed for page 4: %v", err)
	}

	if len(page4.Items) != 0 {
		t.Errorf("Expected 0 items on page 4, got %d", len(page4.Items))
	}

	if page4.HasMore {
		t.Error("Expected HasMore to be false for page 4")
	}
}

func TestGetPagedDirectoryListingWithDefaultPageSize(t *testing.T) {
	// Create mock directory listing with more than default page size
	mockFiles := make([]msg.Finfo, 1500)
	for i := 0; i < 1500; i++ {
		mockFiles[i] = &mockFileInfo{
			name: "file" + string(rune(i+'0')) + ".txt",
			size: int64(i * 10),
			path: "/test/very_large_directory/file" + string(rune(i+'0')) + ".txt",
		}
	}

	// Cache the directory listing
	dirPath := "/test/very_large_directory"
	CacheDirList(dirPath, mockFiles)

	// Test with default page size (0)
	page, err := getPagedDirectoryListing(context.Background(), dirPath, 1, 0)
	if err != nil {
		t.Errorf("getPagedDirectoryListing failed with default page size: %v", err)
	}

	if page.PageSize != DefaultPageSize {
		t.Errorf("Expected page size %d, got %d", DefaultPageSize, page.PageSize)
	}

	if len(page.Items) != DefaultPageSize {
		t.Errorf("Expected %d items, got %d", DefaultPageSize, len(page.Items))
	}
}

func TestGetPagedDirectoryListingWithMaxPageSize(t *testing.T) {
	// Create mock directory listing
	mockFiles := make([]msg.Finfo, 150)
	for i := 0; i < 150; i++ {
		mockFiles[i] = &mockFileInfo{
			name: "file" + string(rune(i+'0')) + ".txt",
			size: int64(i * 10),
			path: "/test/directory/file" + string(rune(i+'0')) + ".txt",
		}
	}

	// Cache the directory listing
	dirPath := "/test/directory"
	CacheDirList(dirPath, mockFiles)

	// Test with page size larger than max
	page, err := getPagedDirectoryListing(context.Background(), dirPath, 1, MaxPageSize+100)
	if err != nil {
		t.Errorf("getPagedDirectoryListing failed with large page size: %v", err)
	}

	if page.PageSize != MaxPageSize {
		t.Errorf("Expected page size %d, got %d", MaxPageSize, page.PageSize)
	}
}

func TestGetPagedDirectoryListingInvalidPage(t *testing.T) {
	// Create mock directory listing
	mockFiles := make([]msg.Finfo, 10)
	for i := 0; i < 10; i++ {
		mockFiles[i] = &mockFileInfo{
			name: "file" + string(rune(i+'0')) + ".txt",
			size: int64(i * 10),
			path: "/test/small_directory/file" + string(rune(i+'0')) + ".txt",
		}
	}

	// Cache the directory listing
	dirPath := "/test/small_directory"
	CacheDirList(dirPath, mockFiles)

	// Test with invalid page number (0)
	page, err := getPagedDirectoryListing(context.Background(), dirPath, 0, 10)
	if err != nil {
		t.Errorf("getPagedDirectoryListing failed with invalid page: %v", err)
	}

	// Should default to page 1
	if page.Page != 1 {
		t.Errorf("Expected page 1 for invalid page 0, got %d", page.Page)
	}
}

func TestGetPagedDirectoryListingEmptyDirectory(t *testing.T) {
	// Create empty directory listing
	var mockFiles []msg.Finfo

	// Cache the directory listing
	dirPath := "/test/empty_directory"
	CacheDirList(dirPath, mockFiles)

	// Test first page of empty directory
	page, err := getPagedDirectoryListing(context.Background(), dirPath, 1, 50)
	if err != nil {
		t.Errorf("getPagedDirectoryListing failed for empty directory: %v", err)
	}

	if page.TotalCount != 0 {
		t.Errorf("Expected total count 0, got %d", page.TotalCount)
	}

	if len(page.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(page.Items))
	}

	if page.HasMore {
		t.Error("Expected HasMore to be false for empty directory")
	}
}