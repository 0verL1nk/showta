package webdav

import (
	"time"

	"overlink.top/app/storage"
	"overlink.top/app/system/model"
	"overlink.top/app/system/msg"
)

// mockStorage implements storage.Storage interface for testing
type mockStorage struct {
	files map[string][]byte
}

func (m *mockStorage) GetConfig() storage.Config {
	return storage.Config{Name: "mock", Direct: true, NoCache: false}
}

func (m *mockStorage) SetData(data model.Storage) {
	// Mock implementation
}

func (m *mockStorage) GetData() *model.Storage {
	return &model.Storage{}
}

func (m *mockStorage) GetExtra() storage.ExtraItem {
	return nil
}

func (m *mockStorage) Mount() error {
	return nil
}

func (m *mockStorage) List(info msg.Finfo) ([]msg.Finfo, error) {
	return nil, nil
}

func (m *mockStorage) Link(info msg.Finfo) (*msg.LinkInfo, error) {
	path := info.GetPath()
	if _, exists := m.files[path]; exists {
		return &msg.LinkInfo{Url: "mock://" + path}, nil
	}
	return nil, nil
}

func (m *mockStorage) AllowCache() bool {
	return true
}

func (m *mockStorage) IsDirect() bool {
	return true
}

// StreamFile implements streaming for mock storage
func (m *mockStorage) StreamFile(ctx interface{}, path string, writer interface{}) error {
	return nil
}

// StreamRange implements range streaming for mock storage
func (m *mockStorage) StreamRange(ctx interface{}, path string, offset, length int64, writer interface{}) error {
	return nil
}

// mockFileInfo implements msg.Finfo interface for testing
type mockFileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
	path    string
}

func (m *mockFileInfo) GetName() string       { return m.name }
func (m *mockFileInfo) GetPath() string       { return m.path }
func (m *mockFileInfo) GetFileId() string     { return "" }
func (m *mockFileInfo) GetSize() int64        { return m.size }
func (m *mockFileInfo) ModTime() time.Time    { return m.modTime }
func (m *mockFileInfo) IsDir() bool           { return m.isDir }
func (m *mockFileInfo) GetRaw() string        { return "" }