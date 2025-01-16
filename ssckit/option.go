package ssckit

import (
	"github.com/vela-public/onekit/zapkit"
)

func Protect(flag bool) func(*Application) {
	return func(app *Application) {
		app.config.Protect = flag
	}
}

func Logger(log *zapkit.Logger) func(*Application) {
	return func(app *Application) {
		app.private.Logger = log
	}
}
