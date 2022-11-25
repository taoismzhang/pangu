package arg

import (
	"github.com/pangum/pangu/app"
)

type action[T argumentType] func(ctx *app.Context, value T) error