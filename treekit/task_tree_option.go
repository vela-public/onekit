package treekit

import "github.com/vela-public/onekit/pipekit"

type TaskTreeOption struct {
	create  *pipekit.Chain[*Process]
	error   *pipekit.Chain[error]
	panic   *pipekit.Chain[error]
	report  *pipekit.Chain[*Task]
	protect bool
}

func NewTaskTreeOption() *TaskTreeOption {
	return &TaskTreeOption{
		protect: false,
		create:  pipekit.NewChain[*Process](),
		error:   pipekit.NewChain[error](),
		panic:   pipekit.NewChain[error](),
		report:  pipekit.NewChain[*Task](),
	}
}

func (tt *TaskTreeOption) Protect(flag bool) {
	tt.protect = flag
}

func (tt *TaskTreeOption) Error(fn func(error)) {
	tt.error.NewHandler(fn)
}

func (tt *TaskTreeOption) Panic(fn func(error)) {
	tt.panic.NewHandler(fn)
}

func (tt *TaskTreeOption) Create(fn func(p *Process)) {
	tt.create.NewHandler(fn)
}

func (tt *TaskTreeOption) Report(fn func(*Task)) {
	tt.report.NewHandler(fn)
}
