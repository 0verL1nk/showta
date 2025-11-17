package webdav

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"overlink.top/app/storage"
	"overlink.top/app/system/model"
	"overlink.top/app/system/msg"
)

// BenchmarkWebDAVCaching benchmarks the performance of the WebDAV caching system
func BenchmarkWebDAVCaching(b *testing.B) {
	// Create test data
	path := "/test/benchmark_file.txt"
	mockInfo := &mockFileInfo{
		name: "benchmark_file.txt",
		size: 1024 * 1024, // 1MB file
		path: "/test/benchmark_file.txt",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cache the file info
		CacheFileInfo(path, mockInfo)

		// Retrieve from cache
		_, err := GetCachedFileInfo(context.Background(), path)
		if err != nil {
			b.Errorf("GetCachedFileInfo failed: %v", err)
		}
	}
}

// BenchmarkWebDAVDirectAccess benchmarks direct access without caching for comparison
func BenchmarkWebDAVDirectAccess(b *testing.B) {
	// Simulate direct access by clearing cache each time
	path := "/test/direct_access_file.txt"
	mockInfo := &mockFileInfo{
		name: "direct_access_file.txt",
		size: 1024 * 1024, // 1MB file
		path: "/test/direct_access_file.txt",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate direct access (cache miss scenario)
		globalFileInfoCache.Clear()

		// Cache the file info
		CacheFileInfo(path, mockInfo)

		// Retrieve from cache
		_, err := GetCachedFileInfo(context.Background(), path)
		if err != nil {
			b.Errorf("GetCachedFileInfo failed: %v", err)
		}
	}
}

// BenchmarkWebDAVStreaming benchmarks the streaming performance
func BenchmarkWebDAVStreaming(b *testing.B) {
	// Create mock storage with test data
	testData := make([]byte, 10*1024*1024) // 10MB of test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	mockStore := &benchmarkMockStorage{
		files: map[string][]byte{
			"/test/stream_benchmark.txt": testData,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		err := mockStore.StreamFile(context.Background(), "/test/stream_benchmark.txt", &buffer)
		if err != nil {
			b.Errorf("StreamFile failed: %v", err)
		}

		// Verify we got the expected amount of data
		if buffer.Len() != len(testData) {
			b.Errorf("Streamed data size mismatch. Expected %d, got %d", len(testData), buffer.Len())
		}
	}
}

// BenchmarkWebDAVRangeStreaming benchmarks the range streaming performance
func BenchmarkWebDAVRangeStreaming(b *testing.B) {
	// Create mock storage with test data
	testData := make([]byte, 10*1024*1024) // 10MB of test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	mockStore := &benchmarkMockStorage{
		files: map[string][]byte{
			"/test/range_stream_benchmark.txt": testData,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		offset := int64(1024 * 1024) // 1MB offset
		length := int64(1024 * 1024) // 1MB length
		err := mockStore.StreamRange(context.Background(), "/test/range_stream_benchmark.txt", offset, length, &buffer)
		if err != nil {
			b.Errorf("StreamRange failed: %v", err)
		}

		// Verify we got the expected amount of data
		if int64(buffer.Len()) != length {
			b.Errorf("Streamed range data size mismatch. Expected %d, got %d", length, buffer.Len())
		}
	}
}

// BenchmarkWebDAVPagedDirectoryListing benchmarks the paged directory listing performance
func BenchmarkWebDAVPagedDirectoryListing(b *testing.B) {
	// Create large mock directory listing
	mockFiles := make([]msg.Finfo, 10000)
	for i := 0; i < 10000; i++ {
		mockFiles[i] = &mockFileInfo{
			name: "file" + string(rune(i+'0')) + ".txt",
			size: int64(i * 100),
			path: "/test/file" + string(rune(i+'0')) + ".txt",
		}
	}

	// Cache the directory listing
	dirPath := "/test/large_benchmark_directory"
	CacheDirList(dirPath, mockFiles)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test paged directory listing
		page, err := getPagedDirectoryListing(context.Background(), dirPath, 1, 100)
		if err != nil {
			b.Errorf("getPagedDirectoryListing failed: %v", err)
		}

		if len(page.Items) != 100 {
			b.Errorf("Expected 100 items, got %d", len(page.Items))
		}
	}
}

// BenchmarkWebDAVCacheExpiration benchmarks cache expiration performance
func BenchmarkWebDAVCacheExpiration(b *testing.B) {
	// Create a cache with a very short TTL for testing
	shortTTL := 1 * time.Microsecond
	testCache := NewFileInfoCache(shortTTL, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := "/test/expiry_benchmark_" + string(rune(i+'0'))

		// Add item to cache
		mockInfo := &mockFileInfo{
			name: "expiry_benchmark_" + string(rune(i+'0')),
			size: int64(i * 1024),
			path: path,
		}
		testCache.SetFile(path, mockInfo)

		// Small delay to ensure expiration
		time.Sleep(shortTTL * 2)

		// Try to retrieve (should trigger expiration cleanup)
		testCache.GetFile(path)
	}
}

// BenchmarkWebDAVSmallFileStreaming benchmarks streaming performance for small files
func BenchmarkWebDAVSmallFileStreaming(b *testing.B) {
	// Create mock storage with small test data (1KB)
	testData := make([]byte, 1*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	mockStore := &benchmarkMockStorage{
		files: map[string][]byte{
			"/test/small_file.txt": testData,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		err := mockStore.StreamFile(context.Background(), "/test/small_file.txt", &buffer)
		if err != nil {
			b.Errorf("StreamFile failed: %v", err)
		}

		// Verify we got the expected amount of data
		if buffer.Len() != len(testData) {
			b.Errorf("Streamed data size mismatch. Expected %d, got %d", len(testData), buffer.Len())
		}
	}
}

// BenchmarkWebDAVLargeFileStreaming benchmarks streaming performance for large files
func BenchmarkWebDAVLargeFileStreaming(b *testing.B) {
	// Create mock storage with large test data (100MB)
	testData := make([]byte, 100*1024*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	mockStore := &benchmarkMockStorage{
		files: map[string][]byte{
			"/test/large_file.txt": testData,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		err := mockStore.StreamFile(context.Background(), "/test/large_file.txt", &buffer)
		if err != nil {
			b.Errorf("StreamFile failed: %v", err)
		}

		// Verify we got the expected amount of data
		if buffer.Len() != len(testData) {
			b.Errorf("Streamed data size mismatch. Expected %d, got %d", len(testData), buffer.Len())
		}
	}
}

// BenchmarkWebDAVCacheHit benchmarks cache hit performance
func BenchmarkWebDAVCacheHit(b *testing.B) {
	path := "/test/cache_hit_file.txt"
	mockInfo := &mockFileInfo{
		name: "cache_hit_file.txt",
		size: 1024,
		path: "/test/cache_hit_file.txt",
	}

	// Pre-cache the file info
	CacheFileInfo(path, mockInfo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Retrieve from cache (should be a hit)
		_, err := GetCachedFileInfo(context.Background(), path)
		if err != nil {
			b.Errorf("GetCachedFileInfo failed: %v", err)
		}
	}
}

// BenchmarkWebDAVCacheMiss benchmarks cache miss performance
func BenchmarkWebDAVCacheMiss(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/test/cache_miss_file_%d.txt", i)

		// Try to retrieve from cache (should be a miss)
		_, err := GetCachedFileInfo(context.Background(), path)
		if err != nil {
			b.Errorf("GetCachedFileInfo failed: %v", err)
		}
	}
}

// BenchmarkWebDAVDirectoryListing benchmarks directory listing performance
func BenchmarkWebDAVDirectoryListing(b *testing.B) {
	// Create medium-sized mock directory listing
	mockFiles := make([]msg.Finfo, 1000)
	for i := 0; i < 1000; i++ {
		mockFiles[i] = &mockFileInfo{
			name: fmt.Sprintf("file_%d.txt", i),
			size: int64(i * 1024),
			path: fmt.Sprintf("/test/dir/file_%d.txt", i),
		}
	}

	// Cache the directory listing
	dirPath := "/test/benchmark_directory"
	CacheDirList(dirPath, mockFiles)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Retrieve directory listing from cache
		_, err := GetCachedDirList(context.Background(), dirPath)
		if err != nil {
			b.Errorf("GetCachedDirList failed: %v", err)
		}
	}
}

// benchmarkMockStorage implements storage.Storage interface for benchmarking
type benchmarkMockStorage struct {
	files map[string][]byte
}

func (m *benchmarkMockStorage) GetConfig() storage.Config {
	return storage.Config{Name: "mock", Direct: true, NoCache: false}
}

func (m *benchmarkMockStorage) SetData(data model.Storage) {
	// Mock implementation
}

func (m *benchmarkMockStorage) GetData() *model.Storage {
	return &model.Storage{}
}

func (m *benchmarkMockStorage) GetExtra() storage.ExtraItem {
	return nil
}

func (m *benchmarkMockStorage) Mount() error {
	return nil
}

func (m *benchmarkMockStorage) List(info msg.Finfo) ([]msg.Finfo, error) {
	return nil, nil
}

func (m *benchmarkMockStorage) Link(info msg.Finfo) (*msg.LinkInfo, error) {
	path := info.GetPath()
	if _, exists := m.files[path]; exists {
		return &msg.LinkInfo{Url: "mock://" + path}, nil
	}
	return nil, nil
}

func (m *benchmarkMockStorage) AllowCache() bool {
	return true
}

func (m *benchmarkMockStorage) IsDirect() bool {
	return true
}

// StreamFile implements streaming for mock storage
func (m *benchmarkMockStorage) StreamFile(ctx context.Context, path string, writer io.Writer) error {
	if data, exists := m.files[path]; exists {
		_, err := bytes.NewReader(data).WriteTo(writer)
		return err
	}
	return nil
}

// StreamRange implements range streaming for mock storage
func (m *benchmarkMockStorage) StreamRange(ctx context.Context, path string, offset, length int64, writer io.Writer) error {
	if data, exists := m.files[path]; exists {
		if offset >= int64(len(data)) {
			return io.EOF
		}
		end := offset + length
		if end > int64(len(data)) {
			end = int64(len(data))
		}
		_, err := bytes.NewReader(data[offset:end]).WriteTo(writer)
		return err
	}
	return nil
}