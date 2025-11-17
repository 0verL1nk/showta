package disk115

import (
	"overlink.top/app/storage"
	"overlink.top/app/system/logic"
	"overlink.top/app/system/model"
	"overlink.top/app/system/msg"
)

type Extra struct {
}

func init() {
	logic.RegisterEngine(func() storage.Storage {
		return &Disk115{}
	})
}

type Disk115 struct {
	model.Storage
	Extra
}

var config = storage.Config{
	Name: "115disk",
}

func (self *Disk115) GetConfig() storage.Config {
	return config
}

func (self *Disk115) AllowCache() bool {
	return !config.NoCache
}

func (self *Disk115) IsDirect() bool {
	return config.Direct
}

func (self *Disk115) GetExtra() storage.ExtraItem {
	return &self.Extra
}

func (self *Disk115) Mount() error {
	return nil
}

func (self *Disk115) List(info msg.Finfo) (list []msg.Finfo, err error) {
	return
}

func (self *Disk115) Link(info msg.Finfo) (*msg.LinkInfo, error) {
	return nil, nil
}
