package get

import (
	"github.com/harluo/boot/internal/application"
	"github.com/harluo/di"
)

type Settings struct {
	di.Get

	Arguments []application.Argument `group:"settings"`
}
