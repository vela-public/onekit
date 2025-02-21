package treekit

import (
	"context"
	"github.com/panjf2000/ants/v2"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/pipekit"
	"sync"
)

type TaskTree struct {
	cache struct {
		mutex sync.RWMutex
		tasks []*Task
	}

	private struct {
		context context.Context
		cancel  context.CancelFunc
		luakit  *luakit.Kit
		protect bool
		threads *ants.Pool
	}

	handler struct {
		Create *pipekit.Chain[*Task]
		Error  *pipekit.Chain[error]
		Panic  *pipekit.Chain[error]
		Report *pipekit.Chain[*Task]
	}
}

func (t *TaskTree) Protect() bool {
	return t.private.protect
}

func (t *TaskTree) Have(eid int64) bool {
	t.cache.mutex.RLock()
	defer t.cache.mutex.RUnlock()
	for _, task := range t.cache.tasks {
		if task.config.ExecID == eid {
			return true
		}
	}
	return false
}

func (t *TaskTree) Context() context.Context {
	return t.private.context
}

func (t *TaskTree) Report(tas *Task) {
	if t.handler.Report != nil {
		t.handler.Report.Invoke(tas)
	}
}
func (t *TaskTree) Error(err error) {
	if t.handler.Error != nil {
		t.handler.Error.Invoke(err)
	}
}

func (t *TaskTree) Panic(err error) {
	if t.handler.Error != nil {
		t.handler.Panic.Invoke(err)
	}
}

func (t *TaskTree) Submit(tas *Task) {
	if t.private.threads == nil {
		return
	}

	err := t.private.threads.Submit(tas.pcall)
	if err != nil {
		t.Error(err)
	}

}

func NewTaskTree(parent context.Context, kit *luakit.Kit, option *TaskTreeOption) *TaskTree {
	ctx, cancel := context.WithCancel(parent)
	tree := &TaskTree{}
	tree.private.context = ctx
	tree.private.cancel = cancel
	tree.private.luakit = kit
	tree.private.protect = option.protect
	tree.handler.Report = option.report
	tree.handler.Error = option.error
	tree.handler.Panic = option.panic

	threads, err := ants.NewPool(128)
	if err != nil {
		tree.Error(err)
		return tree
	}

	tree.private.threads = threads
	return tree
}
