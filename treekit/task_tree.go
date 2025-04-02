package treekit

import (
	"context"
	"github.com/vela-public/onekit/gopool"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/pipe"
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
		threads gopool.Pool
	}

	handler struct {
		Create *pipe.Chain
		Error  *pipe.Chain
		Panic  *pipe.Chain
		Report *pipe.Chain
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
	t.private.threads.CtxGo(t.Context(), tas.pcall)
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
	tree.private.threads = gopool.NewPool("task.tree", int32(128), gopool.NewConfig())
	return tree
}
