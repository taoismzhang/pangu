package main

import (
	`github.com/storezhang/pangu`
)

func validatorDisable() {
	panic(pangu.New(
		pangu.Name("example"),
		pangu.DisableValidate(),
	).Run(newBootstrap))
}
