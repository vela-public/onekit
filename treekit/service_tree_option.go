package treekit

import (
	"github.com/vela-public/onekit/pipe"
)

type MicroServiceOption struct {
	create  *pipe.Chain
	error   *pipe.Chain
	wakeup  *pipe.Chain
	panic   *pipe.Chain
	report  *pipe.Chain
	protect bool
}

func NewMicoServiceOption() *MicroServiceOption {
	return &MicroServiceOption{
		protect: false,
		create:  pipe.NewChain(),
		error:   pipe.NewChain(),
		wakeup:  pipe.NewChain(),
		panic:   pipe.NewChain(),
		report:  pipe.NewChain(),
	}
}

func (mso *MicroServiceOption) Protect(flag bool) {
	mso.protect = flag
}

func (mso *MicroServiceOption) Create(fn func(*Process)) {
	mso.create.NewHandler(fn)
}

func (mso *MicroServiceOption) Error(fn func(error)) {
	mso.error.NewHandler(fn)
}

func (mso *MicroServiceOption) Wakeup(fn func(*Process)) {
	mso.wakeup.NewHandler(fn)
}

func (mso *MicroServiceOption) Panic(fn func(error)) {
	mso.panic.NewHandler(fn)
}

func (mso *MicroServiceOption) Report(fn func(*MsTree)) {
	mso.report.NewHandler(fn)
}
