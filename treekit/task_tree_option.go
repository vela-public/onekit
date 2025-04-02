package treekit

import (
	"github.com/vela-public/onekit/pipe"
)

type TaskTreeOption struct {
	create  *pipe.Chain
	error   *pipe.Chain
	panic   *pipe.Chain
	report  *pipe.Chain
	protect bool
}

func NewTaskTreeOption() *TaskTreeOption {
	return &TaskTreeOption{
		protect: false,
		create:  pipe.NewChain(),
		error:   pipe.NewChain(),
		panic:   pipe.NewChain(),
		report:  pipe.NewChain(),
	}
}

func (tt *TaskTreeOption) Protect(flag bool) {
	tt.protect = flag
}

func (tt *TaskTreeOption) Error(fn func(error)) {
	tt.error.NewHandler(func(v any) {
		if err, ok := v.(error); ok {
			fn(err)
		}
	})
}

func (tt *TaskTreeOption) Panic(fn func(error)) {
	tt.panic.NewHandler(func(v any) {
		if err, ok := v.(error); ok {
			fn(err)
		}
	})
}

func (tt *TaskTreeOption) Create(fn func(p *Process)) {
	tt.create.NewHandler(func(v any) {
		if p, ok := v.(*Process); ok {
			fn(p)
		}
	})
}

func (tt *TaskTreeOption) Report(fn func(*Task)) {
	tt.report.NewHandler(func(v any) {
		if t, ok := v.(*Task); ok {
			fn(t)
		}
	})
}
