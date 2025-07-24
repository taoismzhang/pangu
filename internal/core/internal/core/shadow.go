package core

import (
	"github.com/urfave/cli/v2"
)

type Shadow struct {
	*cli.App
}

func newShadow() (shadow *Shadow) {
	shadow = new(Shadow)
	shadow.App = cli.NewApp()
	// 对于找不到的命令，暂时不做任何处理
	shadow.CommandNotFound = shadow.notfound

	return
}

func (s *Shadow) notfound(_ *cli.Context, _ string) {}
