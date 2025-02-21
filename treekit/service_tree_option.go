package treekit

import (
	"github.com/vela-public/onekit/pipekit"
)

type MicroServiceOption struct {
	create  *pipekit.Chain[*Process]
	error   *pipekit.Chain[error]
	wakeup  *pipekit.Chain[*Process]
	panic   *pipekit.Chain[error]
	report  *pipekit.Chain[*MsTree]
	protect bool
}

func NewMicoServiceOption() *MicroServiceOption {
	return &MicroServiceOption{
		protect: false,
		create:  pipekit.NewChain[*Process](),
		error:   pipekit.NewChain[error](),
		wakeup:  pipekit.NewChain[*Process](),
		panic:   pipekit.NewChain[error](),
		report:  pipekit.NewChain[*MsTree](),
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
