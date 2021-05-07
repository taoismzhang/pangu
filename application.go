package pangu

import (
	`encoding/json`
	`errors`
	`flag`
	`io/fs`
	`io/ioutil`
	`os`
	`path/filepath`
	`reflect`
	`strings`
	`sync`

	`github.com/pelletier/go-toml`
	`github.com/storezhang/glog`
	`github.com/storezhang/gox`
	`github.com/storezhang/pangu/app`
	`github.com/storezhang/pangu/command`
	`github.com/urfave/cli/v2`
	`go.uber.org/dig`
	`gopkg.in/yaml.v3`
)

// Application 应用程序，可以加入两种种类型的程序
// Serve 用于描述应用程序内的服务
// Command 用于描述应用程序内可以被执行的命令
type Application struct {
	// 应用程序本身的配置
	// 徽标
	banner string

	container *dig.Container
}

// New 创建一个应用程序
func New() *Application {
	return &Application{
		container: dig.New(),
	}
}

// AddServe 添加一个服务器到应用程序中
func (a *Application) AddServe(serve Serve) error {
	return a.container.Invoke(func(cmd *command.Serve) {
		cmd.Add(serve)
	})
}

// AddCommand 添加一个可以被执行的命令到应用程序中
func (a *Application) AddCommand(command app.Command) error {
	return a.container.Invoke(func(a *cli.App) {
		a.Commands = append(a.Commands, &cli.Command{
			Name:    command.GetName(),
			Aliases: command.GetAliases(),
			Usage:   command.GetUsage(),
			Action: func(ctx *cli.Context) error {
				return command.Run(app.NewContext(ctx))
			},
		})
	})
}

// AddArgs 添加参数
func (a *Application) AddArgs(arg app.Arg) error {
	return a.container.Invoke(func(app *cli.App) {
		switch arg.GetValue().(type) {
		case string:
			app.Flags = append(app.Flags, &cli.StringFlag{
				Name:        arg.GetName(),
				Aliases:     arg.GetAliases(),
				Usage:       arg.GetUsage(),
				Value:       arg.GetValue().(string),
				DefaultText: arg.GetDefaultText(),
			})
		case bool:
			app.Flags = append(app.Flags, &cli.BoolFlag{
				Name:        arg.GetName(),
				Aliases:     arg.GetAliases(),
				Usage:       arg.GetUsage(),
				Value:       arg.GetValue().(bool),
				DefaultText: arg.GetDefaultText(),
			})
		case int:
			app.Flags = append(app.Flags, &cli.IntFlag{
				Name:        arg.GetName(),
				Aliases:     arg.GetAliases(),
				Usage:       arg.GetUsage(),
				Value:       arg.GetValue().(int),
				DefaultText: arg.GetDefaultText(),
			})
		}
	})
}

// AddMigration 添加一个升级脚本到系统中
func (a *Application) AddMigration(migration fs.FS) error {
	return a.container.Invoke(func(cmd *command.Serve) {
		cmd.AddMigration(migration)
	})
}

// SetBanner 设置徽标
func (a *Application) SetBanner(banner string) {
	a.banner = banner
}

func (a *Application) Set(constructor interface{}) (err error) {
	return a.container.Provide(constructor)
}

func (a *Application) Sets(constructors ...interface{}) (err error) {
	for _, constructor := range constructors {
		if err = a.container.Provide(constructor); nil != err {
			return
		}
	}

	return
}

func (a *Application) Get(function interface{}) error {
	return a.container.Invoke(function)
}

// Run 启动应用程序
func (a *Application) Run(bootstrap func(*Application) Bootstrap) (err error) {
	if err = a.addProvides(); nil != err {
		return
	}

	// 添加启动器到依赖关系中
	if err = a.container.Provide(bootstrap); nil != err {
		return
	}

	// 增加内置的命令及参数
	if err = a.addInternalFlags(); nil != err {
		return
	}
	if err = a.addInternalCommands(); nil != err {
		return
	}

	// 加载用户启动器并做好配置
	if err = a.container.Invoke(func(bootstrap Bootstrap) error {
		return bootstrap.Setup()
	}); nil != err {
		return
	}

	// 启动应用
	err = a.container.Invoke(func(startup *cli.App) error {
		return startup.Run(os.Args)
	})

	return
}

// GetConfig 取得解析后的配置
func (a *Application) GetConfig(config interface{}) (err error) {
	var (
		once sync.Once
		path *string
	)

	// 参数不允许重复定义，只能执行一次
	once.Do(func() {
		path = flag.String("conf", "./conf/application.yaml", "指定配置文件路径")
		flag.StringVar(path, "c", *path, "指定配置文件路径")
		flag.StringVar(path, "config", *path, "指定配置文件路径")
		flag.StringVar(path, "configuration", *path, "指定配置文件路径")
		flag.Parse()
	})

	if reflect.ValueOf(config).Kind() == reflect.Ptr {
		err = a.loadConfig(config, *path)
	} else {
		err = a.loadConfig(&config, *path)
	}

	return
}

func (a *Application) addInternalCommands() error {
	return a.container.Invoke(func(startup *cli.App, serve *command.Serve, version *command.Version) {
		startup.Commands = append(startup.Commands, &cli.Command{
			Name:    serve.GetName(),
			Aliases: serve.GetAliases(),
			Usage:   serve.GetUsage(),
			Action: func(ctx *cli.Context) error {
				return serve.Run(app.NewContext(ctx))
			},
		})
		startup.Commands = append(startup.Commands, &cli.Command{
			Name:    version.GetName(),
			Aliases: version.GetAliases(),
			Usage:   version.GetUsage(),
			Action: func(ctx *cli.Context) error {
				return version.Run(app.NewContext(ctx))
			},
		})
	})
}

func (a *Application) addInternalFlags() error {
	return a.container.Invoke(func(startup *cli.App) {
		startup.Flags = append(startup.Flags, &cli.StringFlag{
			Name:        "conf",
			Aliases:     []string{"c", "Config", "configuration"},
			Usage:       "指定配置文件路径",
			Value:       "./conf/application.yaml",
			DefaultText: "./conf/application.yaml",
		})
	})
}

func (a *Application) loadConfig(config interface{}, path string) (err error) {
	if path, err = a.findConfigFilepath(path); nil != err {
		return
	}

	var data []byte
	if data, err = ioutil.ReadFile(path); nil != err {
		return
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case "yml":
		fallthrough
	case "yaml":
		err = yaml.Unmarshal(data, config)
	case "json":
		err = json.Unmarshal(data, config)
	case "toml":
		err = toml.Unmarshal(data, config)
	default:
		err = yaml.Unmarshal(data, config)
	}

	return
}

func (a *Application) findConfigFilepath(conf string) (path string, err error) {
	path = conf
	if "" == path {
		path = "./application.yml"
	}
	if gox.IsFileExist(path) {
		return
	}

	path = "./application.yaml"
	if gox.IsFileExist(path) {
		return
	}

	path = "./application.toml"
	if gox.IsFileExist(path) {
		return
	}

	path = "./conf/application.yml"
	if gox.IsFileExist(path) {
		return
	}

	path = "./conf/application.yaml"
	if gox.IsFileExist(path) {
		return
	}
	err = errors.New("找不到配置文件")

	return
}

func (a *Application) addProvides() (err error) {
	if err = a.Set(glog.NewLogger); nil != err {
		return
	}
	if err = a.Set(gox.NewSnowflake); nil != err {
		return
	}
	if err = a.Set(NewResty); nil != err {
		return
	}
	if err = a.Sets(NewRedis, NewXormEngine, NewElasticsearch); nil != err {
		return
	}
	if err = a.Sets(command.NewServe, command.NewVersion); nil != err {
		return
	}

	if err = a.Sets(appName, appVersion, buildVersion, buildTime, scmRevision, scmBranch, goVersion); nil != err {
		return
	}

	if err = a.Sets(newApp, newZapLogger); nil != err {
		return
	}

	// 注入自身
	if err = a.container.Provide(func() *Application {
		return a
	}); nil != err {
		return
	}

	return
}
