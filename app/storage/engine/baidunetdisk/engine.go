package baidunetdisk

import (
	"context"
	"io"
	"overlink.top/app/storage"
	"overlink.top/app/system/logic"
	"overlink.top/app/system/model"
	"overlink.top/app/system/msg"
)

type Extra struct {
}

func init() {
	logic.RegisterEngine(func() storage.Storage {
		return &Baidunetdisk{}
	})
}

type Baidunetdisk struct {
	model.Storage
	Extra
}

var config = storage.Config{
	Name: "baidunetdisk",
}

func (self *Baidunetdisk) GetConfig() storage.Config {
	return config
}

func (self *Baidunetdisk) AllowCache() bool {
	return !config.NoCache
}

func (self *Baidunetdisk) IsDirect() bool {
	return config.Direct
}

func (self *Baidunetdisk) GetExtra() storage.ExtraItem {
	return &self.Extra
}

func (self *Baidunetdisk) Mount() error {
	return nil
}

func (self *Baidunetdisk) List(info msg.Finfo) (list []msg.Finfo, err error) {
	return
}

func (self *Baidunetdisk) Link(info msg.Finfo) (*msg.LinkInfo, error) {
	return nil, nil
}

// StreamFile streams a file directly to the writer
func (self *Baidunetdisk) StreamFile(ctx context.Context, rpath string, writer io.Writer) error {
	return storage.DefaultStreamFile(ctx, self, rpath, writer)
}

// StreamRange streams a file range directly to the writer
func (self *Baidunetdisk) StreamRange(ctx context.Context, rpath string, offset, length int64, writer io.Writer) error {
	return storage.DefaultStreamRange(ctx, self, rpath, offset, length, writer)
}
