package param

import (
	"path/filepath"

	"github.com/goexl/env"
	"github.com/goexl/mengpo"
	"github.com/goexl/xiren"
	"github.com/pangum/pangu/internal/callback/getter"
	"github.com/pangum/pangu/internal/config"
	"github.com/pangum/pangu/internal/constant"
	"github.com/pangum/pangu/internal/internal"
	"github.com/pangum/pangu/internal/runtime"
)

type Config struct {
	Paths      []string
	Extensions []string

	Default  bool
	Validate bool
	Nullable bool

	Tag               *Tag
	EnvironmentGetter getter.Environment
	Environments      internal.Environments
	Watcher           config.Watcher
	Loader            config.Loader
}

func newConfig() *Config {
	return &Config{
		Paths: []string{
			constant.ApplicationName,
			filepath.Join(constant.ConfigDir, constant.ApplicationName),
			filepath.Join(constant.ConfigConfDir, constant.ApplicationName),
			filepath.Join(constant.ConfigConfigurationDir, constant.ApplicationName),
		},
		Extensions: []string{
			constant.ExtensionYml,
			constant.ExtensionYaml,
			constant.ExtensionToml,
			constant.ExtensionJson,
			constant.ExtensionXml,
		},
		Default:  true,
		Validate: true,
		Nullable: true,

		Tag:               NewTag(),
		EnvironmentGetter: env.Get,
		Environments:      make(internal.Environments, 0),
	}
}

func (c *Config) Fill(path string, config runtime.Pointer) (err error) {
	if le := c.Loader.Load(path, config); nil != le {
		err = le
	} else if c.Default {
		// 处理默认值，此处逻辑不能往前，原因
		// 如果对象里面包含指针，那么只能在包含指针的结构体被解析后才能去设置默认值，不然指针将被会设置成nil
		err = mengpo.New().Tag(c.Tag.Default).Build().Set(config)
	} else if c.Validate {
		// 数据验证
		err = xiren.Struct(config)
	}

	return
}
