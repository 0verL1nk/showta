package webdav

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"overlink.top/app/storage"
	"overlink.top/app/system/model"
	"overlink.top/app/system/msg"
)

// streamingMockStorage implements storage.Storage interface for testing streaming
type streamingMockStorage struct {
	files map[string][]byte
}

func (m *streamingMockStorage) GetConfig() storage.Config {
	return storage.Config{Name: "mock", Direct: true, NoCache: false}
}

func (m *streamingMockStorage) SetData(data model.Storage) {
	// Mock implementation
}

func (m *streamingMockStorage) GetData() *model.Storage {
	return &model.Storage{}
}

func (m *streamingMockStorage) GetExtra() storage.ExtraItem {
	return nil
}

func (m *streamingMockStorage) Mount() error {
	return nil
}

func (m *streamingMockStorage) List(info msg.Finfo) ([]msg.Finfo, error) {
	return nil, nil
}

func (m *streamingMockStorage) Link(info msg.Finfo) (*msg.LinkInfo, error) {
	path := info.GetPath()
	if _, exists := m.files[path]; exists {
		return &msg.LinkInfo{Url: "mock://" + path}, nil
	}
	return nil, os.ErrNotExist
}

func (m *streamingMockStorage) AllowCache() bool {
	return true
}

func (m *streamingMockStorage) IsDirect() bool {
	return true
}

// StreamFile implements streaming for mock storage
func (m *streamingMockStorage) StreamFile(ctx context.Context, path string, writer io.Writer) error {
	if data, exists := m.files[path]; exists {
		_, err := bytes.NewReader(data).WriteTo(writer)
		return err
	}
	return os.ErrNotExist
}

// StreamRange implements range streaming for mock storage
func (m *streamingMockStorage) StreamRange(ctx context.Context, path string, offset, length int64, writer io.Writer) error {
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
	return os.ErrNotExist
}

func TestStreamFile(t *testing.T) {
	// Create mock storage with test data
	testData := []byte("This is test data for streaming functionality")
	mockStore := &streamingMockStorage{
		files: map[string][]byte{
			"/test/file.txt": testData,
		},
	}

	// Test streaming a file
	var buffer bytes.Buffer
	err := mockStore.StreamFile(context.Background(), "/test/file.txt", &buffer)
	if err != nil {
		t.Errorf("StreamFile failed: %v", err)
	}

	// Verify the streamed data
	result := buffer.Bytes()
	if !bytes.Equal(result, testData) {
		t.Errorf("Streamed data mismatch. Expected %d bytes, got %d bytes", len(testData), len(result))
	}
}

func TestStreamRange(t *testing.T) {
	// Create mock storage with test data
	testData := []byte("This is test data for range streaming functionality")
	mockStore := &streamingMockStorage{
		files: map[string][]byte{
			"/test/range_file.txt": testData,
		},
	}

	// Test streaming a range of data
	var buffer bytes.Buffer
	offset := int64(5)
	length := int64(4)
	err := mockStore.StreamRange(context.Background(), "/test/range_file.txt", offset, length, &buffer)
	if err != nil {
		t.Errorf("StreamRange failed: %v", err)
	}

	// Verify the streamed data
	expected := testData[offset : offset+length]
	result := buffer.Bytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Streamed range data mismatch. Expected '%s', got '%s'", string(expected), string(result))
	}
}

func TestStreamFileNotFound(t *testing.T) {
	// Create mock storage without the requested file
	mockStore := &streamingMockStorage{
		files: map[string][]byte{},
	}

	// Test streaming a non-existent file
	var buffer bytes.Buffer
	err := mockStore.StreamFile(context.Background(), "/nonexistent/file.txt", &buffer)
	if err == nil {
		t.Error("Expected error when streaming non-existent file")
	}
	if err != os.ErrNotExist {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestStreamRangeOutOfBounds(t *testing.T) {
	// Create mock storage with test data
	testData := []byte("Small file")
	mockStore := &streamingMockStorage{
		files: map[string][]byte{
			"/test/small_file.txt": testData,
		},
	}

	// Test streaming a range that exceeds file bounds
	var buffer bytes.Buffer
	offset := int64(5)
	length := int64(100) // Much larger than remaining data
	err := mockStore.StreamRange(context.Background(), "/test/small_file.txt", offset, length, &buffer)
	if err != nil {
		t.Errorf("StreamRange failed unexpectedly: %v", err)
	}

	// Should have streamed only the remaining data
	expected := testData[offset:]
	result := buffer.Bytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Streamed range data mismatch. Expected '%s', got '%s'", string(expected), string(result))
	}
}