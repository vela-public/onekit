package taskit

import (
	"github.com/vela-public/onekit/pipekit"
)

type Options struct {
	Create  *pipekit.Chain[*Service]
	Error   *pipekit.Chain[error]
	Wakeup  *pipekit.Chain[*Service]
	Panic   *pipekit.Chain[error]
	Report  *pipekit.Chain[*Tree]
	Protect bool
}

func OnCreate(v any) func(*Options) {
	return func(opts *Options) {
		opts.Create.NewHandler(v)
	}
}

func OnError[T any](v T) func(*Options) {
	return func(opts *Options) {
		opts.Error.NewHandler(v)
	}
}

func OnWakeup(v any) func(*Options) {
	return func(opts *Options) {
		opts.Wakeup.NewHandler(v)
	}
}

func OnPanic(v any) func(*Options) {
	return func(opts *Options) {
		opts.Panic.NewHandler(v)
	}
}

func Protect(flag bool) func(*Options) {
	return func(opts *Options) {
		opts.Protect = flag
	}
}

func Report(v any) func(*Options) {
	return func(opts *Options) {
		opts.Report.NewHandler(v)
	}
}
