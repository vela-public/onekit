package treekit

import (
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/noop"
	"github.com/vela-public/onekit/pipe"
)

type MicroServiceOption struct {
	create  *pipe.Chain
	error   *pipe.Chain
	debug   *pipe.LazyChain[string]
	wakeup  *pipe.Chain
	panic   *pipe.Chain
	report  *pipe.Chain
	protect bool
}

func DefaultServiceOption() *MicroServiceOption {
	log := noop.NewLogger(2)
	say := func(v string) {
		log.Infof("%s", v)
	}

	err := func(err error) {
		log.Errorf("%v", err)
	}

	option := NewMicoServiceOption()
	option.Protect(true)
	option.Protect(false)
	option.Debug(say)
	option.Error(err)
	option.Panic(err)
	option.Report(func(ms *MsTree) {
		say(cast.B2S(ms.Doc()))
	})

	return option
}

func NewMicoServiceOption() *MicroServiceOption {
	return &MicroServiceOption{
		protect: false,
		create:  pipe.NewChain(),
		error:   pipe.NewChain(),
		debug:   pipe.NewLazyChain[string](),
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

func (mso *MicroServiceOption) Debug(fn func(string)) {
	mso.debug.NewHandler(fn)
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
