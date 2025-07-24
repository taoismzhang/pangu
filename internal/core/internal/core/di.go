package core

import (
	"github.com/harluo/di"
)

func init() {
	di.New().Instance().Put(
		newShell,
		newShadow,
	).Build().Apply()
}
