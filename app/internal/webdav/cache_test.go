package webdav

import (
	"fmt"
	"testing"
	"time"

	"overlink.top/app/system/msg"
)

func TestWebDAVCaching(t *testing.T) {
	// Test cache functionality
	path := "/test/file.txt"

	// Test file cache
	mockInfo := &mockFileInfo{
		name: "file.txt",
		size: 1024,
		path: "/test/file.txt",
	}

	// Create a cache with 5 minute TTL
	cache := NewFileInfoCache(5*time.Minute, 1000)

	// Cache the file info
	cache.SetFile(path, mockInfo)

	// Retrieve from cache
	cachedInfo, found := cache.GetFile(path)
	if !found {
		t.Error("Expected file to be found in cache")
	}

	// Verify the cached info matches the original
	if cachedInfo.GetName() != mockInfo.GetName() {
		t.Errorf("Expected name '%s', got '%s'", mockInfo.GetName(), cachedInfo.GetName())
	}
	if cachedInfo.GetSize() != mockInfo.GetSize() {
		t.Errorf("Expected size %d, got %d", mockInfo.GetSize(), cachedInfo.GetSize())
	}
	if cachedInfo.GetPath() != mockInfo.GetPath() {
		t.Errorf("Expected path '%s', got '%s'", mockInfo.GetPath(), cachedInfo.GetPath())
	}

	// Test non-existent file
	_, found = cache.GetFile("/non/existent/file.txt")
	if found {
		t.Error("Expected non-existent file to not be found in cache")
	}
}

func TestWebDAVCacheExpiration(t *testing.T) {
	// Test cache expiration functionality
	path := "/test/expiring_file.txt"

	// Create a cache with a very short TTL for testing
	shortTTL := 100 * time.Millisecond
	testCache := NewFileInfoCache(shortTTL, 1000)

	// Add item to cache
	mockInfo := &mockFileInfo{
		name: "expiring_file.txt",
		size: 2048,
		path: "/test/expiring_file.txt",
	}
	testCache.SetFile(path, mockInfo)

	// Item should be found immediately
	_, found := testCache.GetFile(path)
	if !found {
		t.Error("Expected item to be found in cache immediately")
	}

	// Wait for expiration
	time.Sleep(shortTTL * 2)

	// Item should no longer be found (expired)
	_, found = testCache.GetFile(path)
	if found {
		t.Error("Expected item to be expired and not found in cache")
	}
}

func TestWebDAVCacheInvalidate(t *testing.T) {
	// Test cache invalidation functionality
	path := "/test/invalidate_file.txt"

	// Create a cache
	testCache := NewFileInfoCache(5*time.Minute, 1000)

	// Add item to cache
	mockInfo := &mockFileInfo{
		name: "invalidate_file.txt",
		size: 4096,
		path: "/test/invalidate_file.txt",
	}
	testCache.SetFile(path, mockInfo)

	// Item should be found
	_, found := testCache.GetFile(path)
	if !found {
		t.Error("Expected item to be found in cache")
	}

	// Invalidate the cache entry
	testCache.Invalidate(path)

	// Item should no longer be found
	_, found = testCache.GetFile(path)
	if found {
		t.Error("Expected item to be invalidated and not found in cache")
	}
}

func TestWebDAVCacheClear(t *testing.T) {
	// Test cache clearing functionality
	testCache := NewFileInfoCache(5*time.Minute, 1000)

	// Add multiple items to cache
	for i := 0; i < 5; i++ {
		path := fmt.Sprintf("/test/clear_%d.txt", i)
		mockInfo := &mockFileInfo{
			name: fmt.Sprintf("clear_%d.txt", i),
			size: int64(i * 1024),
			path: path,
		}
		testCache.SetFile(path, mockInfo)
	}

	// Verify items are in cache
	for i := 0; i < 5; i++ {
		path := fmt.Sprintf("/test/clear_%d.txt", i)
		_, found := testCache.GetFile(path)
		if !found {
			t.Errorf("Expected item %s to be found in cache", path)
		}
	}

	// Clear the cache
	testCache.Clear()

	// Verify items are no longer in cache
	for i := 0; i < 5; i++ {
		path := fmt.Sprintf("/test/clear_%d.txt", i)
		_, found := testCache.GetFile(path)
		if found {
			t.Errorf("Expected item %s to be cleared from cache", path)
		}
	}
}

func TestWebDAVCacheInvalidatePattern(t *testing.T) {
	// Test cache invalidation by pattern functionality
	testCache := NewFileInfoCache(5*time.Minute, 1000)

	// Add items to cache
	paths := []string{"/test/file1.txt", "/test/file2.txt", "/other/file3.txt"}
	names := []string{"file1.txt", "file2.txt", "file3.txt"}

	for i, path := range paths {
		mockInfo := &mockFileInfo{
			name: names[i],
			size: int64((i + 1) * 1024),
			path: path,
		}
		testCache.SetFile(path, mockInfo)
	}

	// Verify all items are in cache
	for _, path := range paths {
		_, found := testCache.GetFile(path)
		if !found {
			t.Errorf("Expected item %s to be found in cache", path)
		}
	}

	// Invalidate items matching pattern "/test/"
	testCache.InvalidatePattern("/test/")

	// Check if items matching pattern were invalidated
	_, found1Ok := testCache.GetFile("/test/file1.txt")
	_, found2Ok := testCache.GetFile("/test/file2.txt")
	_, found3Ok := testCache.GetFile("/other/file3.txt")

	if found1Ok {
		t.Error("Expected items matching pattern '/test/' to be invalidated")
	}
	if found2Ok {
		t.Error("Expected items matching pattern '/test/' to be invalidated")
	}
	if !found3Ok {
		t.Error("Expected items not matching pattern to remain in cache")
	}
}

func TestWebDAVDirectoryListingCache(t *testing.T) {
	// Test directory listing cache functionality
	dirPath := "/test/directory"

	// Create mock directory listing
	mockList := []msg.Finfo{
		&mockFileInfo{name: "file1.txt", size: 100, path: "/test/directory/file1.txt"},
		&mockFileInfo{name: "file2.txt", size: 200, path: "/test/directory/file2.txt"},
		&mockFileInfo{name: "subdir", size: 0, path: "/test/directory/subdir", isDir: true},
	}

	// Create a cache
	testCache := NewFileInfoCache(5*time.Minute, 1000)

	// Cache the directory listing
	testCache.SetDirList(dirPath, mockList)

	// Retrieve from cache
	cachedList, found := testCache.GetDirList(dirPath)
	if !found {
		t.Error("Expected directory listing to be found in cache")
	}

	// Verify the cached list matches the original
	if len(cachedList) != 3 {
		t.Errorf("Expected 3 items in directory list, got %d", len(cachedList))
	}

	// Verify the contents
	if cachedList[0].GetName() != "file1.txt" {
		t.Errorf("Expected first item to be 'file1.txt', got '%s'", cachedList[0].GetName())
	}
	if cachedList[1].GetSize() != 200 {
		t.Errorf("Expected second item size to be 200, got %d", cachedList[1].GetSize())
	}
	if !cachedList[2].IsDir() {
		t.Error("Expected third item to be a directory")
	}
}

func TestWebDAVCacheSizeLimit(t *testing.T) {
	// Test cache size limiting functionality
	ttl := 5 * time.Minute
	maxSize := 3 // Very small cache for testing
	testCache := NewFileInfoCache(ttl, maxSize)

	// Add more items than the cache can hold
	for i := 0; i < 5; i++ {
		path := fmt.Sprintf("/test/size_limit_%d.txt", i)
		mockInfo := &mockFileInfo{
			name: fmt.Sprintf("size_limit_%d.txt", i),
			size: int64(i * 1024),
			path: path,
		}
		testCache.SetFile(path, mockInfo)
	}

	// Cache should only contain the most recent items (last 3)
	// The first 2 items should have been evicted
	for i := 0; i < 2; i++ {
		path := fmt.Sprintf("/test/size_limit_%d.txt", i)
		_, found := testCache.GetFile(path)
		if found {
			t.Errorf("Expected item %s to be evicted from cache due to size limit", path)
		}
	}

	// The last 3 items should still be in cache
	for i := 2; i < 5; i++ {
		path := fmt.Sprintf("/test/size_limit_%d.txt", i)
		_, found := testCache.GetFile(path)
		if !found {
			t.Errorf("Expected item %s to be in cache", path)
		}
	}
}