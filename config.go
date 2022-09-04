package pangu

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"

	"github.com/drone/envsubst"
	"github.com/goexl/exc"
	"github.com/goexl/gfx"
	"github.com/goexl/gox/field"
	"github.com/goexl/mengpo"
	"github.com/goexl/xiren"
	"github.com/urfave/cli/v2"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

const (
	configDir        = `config`
	confDir          = `conf`
	configurationDir = `configuration`
)

// Config 配置处理器
type Config struct {
	// 路径
	path string
	// 原始数据
	data []byte
	// 选项
	options *options
}

func newConfig(options *options) *Config {
	return &Config{
		options: options,
	}
}

func (c *Config) Load(config interface{}, opts ...configOption) (err error) {
	for _, opt := range opts {
		opt.applyConfig(c.options.configOptions)
	}
	err = c.loadConfig(config)

	return
}

func (c *Config) loadConfig(config interface{}) (err error) {
	if path, pe := c.configFilepath(c.path); nil != pe {
		if c.options.configOptions.nullable { // 可以不需要配置文件
			c.data = []byte(``)
		} else { // 如果配置为必须要配置文件，抛出错误
			err = pe
		}
	} else if c.loadable() {
		c.path = path
		c.data, err = os.ReadFile(path)
	}
	if nil != err {
		return
	}

	// 处理环境变量，不能修改原始数据，复制一份原始数据做修改
	var _data string
	if _data, err = envsubst.Eval(string(c.data), c.options.environmentGetter); nil != err {
		return
	}

	switch strings.ToLower(filepath.Ext(c.path)) {
	case ymlExt:
		fallthrough
	case yamlExt:
		err = yaml.Unmarshal([]byte(_data), config)
	case jsonExt:
		err = json.Unmarshal([]byte(_data), config)
	case tomlExt:
		err = toml.Unmarshal([]byte(_data), config)
	case xmlExt:
		err = xml.Unmarshal([]byte(_data), config)
	default:
		err = yaml.Unmarshal([]byte(_data), config)
	}
	if nil != err {
		return
	}

	// 处理默认值，此处逻辑不能往前，原因
	// 如果对象里面包含指针，那么只能在包含指针的结构体被解析后才能去设置默认值，不然指针将被会设置成nil
	if c.options.defaults {
		if err = mengpo.Set(config, mengpo.Tag(c.options.tag.defaults)); nil != err {
			return
		}
	}

	// 验证数据
	if c.options.validates {
		err = xiren.Struct(config)
	}

	return
}

func (c *Config) configFilepath(conf string) (path string, err error) {
	gfxOptions := gfx.NewExistsOptions(
		gfx.Paths(c.options.paths...),
		gfx.Extensions(c.options.extensions...),
	)
	// 如果配置了应用名称，可以使用应用名称的配置文件
	if defaultName != Name {
		gfxOptions = append(gfxOptions, gfx.Paths(
			Name,
			filepath.Join(configDir, Name),
			filepath.Join(confDir, Name),
			filepath.Join(configurationDir, Name),
		))
	}

	if final, exists := gfx.Exists(conf, gfxOptions...); exists {
		path = final
	} else {
		err = exc.NewField(`找不到配置文件`, field.String(`path`, final))
	}

	return
}

func (c *Config) bind(shell *cli.App, shadow *cli.App) {
	configFlag := &cli.StringFlag{
		Name:        `config`,
		Aliases:     []string{`c`, `conf`, `configuration`},
		Value:       `./conf/application.yaml`,
		Usage:       `指定配置文件路径`,
		Destination: &c.path,
	}
	shell.Flags = append(shell.Flags, configFlag)
	shadow.Flags = append(shadow.Flags, configFlag)
}

func (c *Config) loadable() bool {
	return `` == c.path || nil == c.data
}
