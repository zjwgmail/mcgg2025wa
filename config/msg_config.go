package config

import (
	"fmt"
	"go-fission-activity/util/config"
	"go-fission-activity/util/config/source"
	"log"
)

var (
	_msgSet *MsgSetting
)

// MsgSetting 兼容原先的配置结构
type MsgSetting struct {
	MsgSetting *MsgYml `yaml:"msgSetting"`
	callbacks  []func()
}

func (e *MsgSetting) runCallback() {
	for i := range e.callbacks {
		e.callbacks[i]()
	}
}

func (e *MsgSetting) OnChange() {
	e.init()
	log.Println("!!! config change and reload")
}

func (e *MsgSetting) Init() {
	e.init()
	log.Println("!!! config init")
}

func (e *MsgSetting) init() {
	log.Println("change init..........")
	e.runCallback()
}

// MsgSetup 载入配置文件
func MsgSetup(s source.Source, fs ...func()) {
	_msgSet = &MsgSetting{
		MsgSetting: MsgConfig,
		callbacks:  fs,
	}
	var err error
	config.DefaultConfig, err = config.NewConfig(
		config.WithSource(s),
		config.WithEntity(_msgSet),
	)
	if err != nil {
		log.Fatal(fmt.Sprintf("New config object fail: %s", err.Error()))
	}
	_msgSet.Init()
}
