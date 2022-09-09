package arg

import (
	"github.com/pangum/pangu/app"
	"github.com/urfave/cli/v2"
)

var (
	_         = NewInts
	_ app.Arg = (*intsArg)(nil)
)

type intsArg struct {
	*Argument

	destination *[]int
}

// NewInts 创建一个整形数组参数
func NewInts(base *Argument, destination *[]int, values ...int) *intsArg {
	return &intsArg{
		Argument:    base,
		destination: destination,
	}
}

func (i *intsArg) Destination() any {
	return i.destination
}

func (i *intsArg) Flag() (flag app.Flag) {
	isf := &cli.IntSliceFlag{
		Name:        i.Name(),
		Aliases:     i.Aliases(),
		Usage:       i.Usage(),
		DefaultText: i.DefaultText(),
		Required:    i.Required(),
		Hidden:      i.Hidden(),
	}
	if nil != i.Default() {
		isf.Value = cli.NewIntSlice(i.Default().([]int)...)
	}
	if nil != i.Destination() {
		isf.Destination = cli.NewIntSlice(*i.Destination().(*[]int)...)
	}
	flag = isf

	return
}
