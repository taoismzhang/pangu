package pangu

var (
	_ option        = (*optionNamed)(nil)
	_ provideOption = (*optionNamed)(nil)
	_ invokeOption  = (*optionNamed)(nil)
)

type optionNamed struct {
	name string
}

// Named 配置应用名称
func Named(name string) *optionNamed {
	return &optionNamed{
		name: name,
	}
}

func (n *optionNamed) apply(_ *options) {
	Name = n.name
}

func (n *optionNamed) applyProvide(options *provideOptions) {
	options.name = n.name
}

func (n *optionNamed) applyInvoke(options *invokeOptions) {
	options.name = n.name
}
