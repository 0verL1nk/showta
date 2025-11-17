package native

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"overlink.top/app/lib/util"
	"overlink.top/app/storage"
	"overlink.top/app/system/conf"
	"overlink.top/app/system/logic"
	"overlink.top/app/system/model"
	"overlink.top/app/system/msg"
	"strings"
)

type Extra struct {
	RootPath string `json:"root_path" required:"true" tip:"true"`
}

func (self *Extra) GetRootPath() string {
	return self.RootPath
}

func init() {
	logic.RegisterEngine(func() storage.Storage {
		return &Native{}
	})
}

type Native struct {
	model.Storage
	Extra
}

var config = storage.Config{
	Name:    "native",
	Direct:  true,
	NoCache: true,
}

func (self *Native) GetConfig() storage.Config {
	return config
}

func (self *Native) AllowCache() bool {
	return !config.NoCache
}

func (self *Native) IsDirect() bool {
	return config.Direct
}

func (self *Native) GetExtra() storage.ExtraItem {
	return &self.Extra
}

func (self *Native) Mount() error {
	exist, err := util.IsDirExist(self.RootPath)
	if !exist {
		return fmt.Errorf("native mount err: %+v", err)
	}
	return nil
}

func (self *Native) Get(rpath string) (info msg.Finfo, err error) {
	apath := self.getApath(rpath)
	fileinfo, err := os.Stat(apath)
	if err != nil {
		err = errors.New("dir err")
		return
	}

	info = &msg.FileInfo{
		Name:     fileinfo.Name(),
		IsFolder: fileinfo.IsDir(),
		Modified: fileinfo.ModTime(),
		Size:     fileinfo.Size(),
	}
	return
}

func (self *Native) List(info msg.Finfo) (list []msg.Finfo, err error) {
	rpath := info.GetPath()
	apath := self.getApath(rpath)
	dir, err := ioutil.ReadDir(apath)
	if err != nil {
		return
	}

	for _, file := range dir {
		list = append(list, &msg.FileInfo{
			Path:     path.Join("/", rpath, file.Name()),
			Name:     file.Name(),
			Size:     file.Size(),
			Modified: file.ModTime(),
			IsFolder: file.IsDir(),
		})
	}

	return
}

func (self *Native) Link(info msg.Finfo) (*msg.LinkInfo, error) {
	rpath := info.GetPath()
	apath := self.getApath(rpath)
	return &msg.LinkInfo{Url: apath}, nil
}

func (self *Native) getApath(rpath string) string {
	mountPath := self.GetData().MountPath
	subpath := strings.TrimPrefix(rpath, mountPath)
	apath := filepath.Join(self.GetRootPath(), subpath)
	apath = filepath.ToSlash(apath)
	return apath
}

// StreamFile streams a file directly to the writer
func (self *Native) StreamFile(ctx context.Context, rpath string, writer io.Writer) error {
	apath := self.getApath(rpath)

	// Check if file exists and is not a directory
	fileInfo, err := os.Stat(apath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return errors.New("cannot stream a directory")
	}

	// Open file for reading
	file, err := os.Open(apath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use chunked transfer encoding for large files
	// For HTTP responses, Go automatically handles chunked transfer encoding
	// when the Content-Length is not set and the connection is HTTP/1.1+

	// Use configured buffer size or default to 64KB
	bufferSize := 64 * 1024
	if conf.AppConf.WebDAV.BufferSize > 0 {
		bufferSize = conf.AppConf.WebDAV.BufferSize
	}
	buf := make([]byte, bufferSize)
	_, err = io.CopyBuffer(writer, file, buf)
	return err
}

// StreamRange streams a file range directly to the writer
func (self *Native) StreamRange(ctx context.Context, rpath string, offset, length int64, writer io.Writer) error {
	apath := self.getApath(rpath)

	// Check if file exists and is not a directory
	fileInfo, err := os.Stat(apath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return errors.New("cannot stream a directory")
	}

	// Open file for reading
	file, err := os.Open(apath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Seek to the specified offset
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	// Create a limited reader for the specified length
	limitedReader := io.LimitReader(file, length)

	// Copy the limited content to writer with buffered copying
	// Use configured buffer size or default to 64KB
	bufferSize := 64 * 1024
	if conf.AppConf.WebDAV.BufferSize > 0 {
		bufferSize = conf.AppConf.WebDAV.BufferSize
	}
	buf := make([]byte, bufferSize)
	_, err = io.CopyBuffer(writer, limitedReader, buf)
	return err
}
