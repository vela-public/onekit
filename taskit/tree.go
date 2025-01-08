package taskit

import (
	"context"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/pipekit"
	"sync"
)

type Tree struct {
	cache struct {
		mutex sync.RWMutex
		data  []*task
	}

	handler struct {
		Report *pipekit.Chain[*Tree]
		Create *pipekit.Chain[*Service]
		Error  *pipekit.Chain[error]
		Wakeup *pipekit.Chain[*Service]
		Panic  *pipekit.Chain[error]
	}

	private struct {
		protect bool
		context context.Context
		cancel  context.CancelFunc
		parent  *luakit.Kit
		error   error
	}
}

func (t *Tree) Protect() bool {
	return t.private.protect
}

func (t *Tree) NewKit() *luakit.Kit {
	return t.private.parent.Clone()
}

func (t *Tree) Context() context.Context {
	return t.private.context
}

func (t *Tree) push(tas *task) {
	t.cache.mutex.Lock()
	defer t.cache.mutex.Unlock()
	t.cache.data = append(t.cache.data, tas)
}

func (t *Tree) find(key string) (*task, bool) {
	t.cache.mutex.RLock()
	defer t.cache.mutex.RUnlock()

	sz := len(t.cache.data)
	if sz == 0 {
		return nil, false
	}

	for i := 0; i < sz; i++ {
		tas := t.cache.data[i]
		if tas.config.Key == key {
			return tas, true
		}
	}

	return nil, false
}

func (t *Tree) RemoveByID(ids []int64) {
	t.cache.mutex.Lock()
	defer t.cache.mutex.Unlock()
	var dat []*task
	for i, tas := range t.cache.data {
		if !libkit.In[int64](ids, tas.ID()) {
			dat = append(dat, t.cache.data[i])
			continue
		}
		_ = tas.Close()
	}
	t.cache.data = dat
}

func NewTree(parent context.Context, kit *luakit.Kit, options ...func(*Options)) *Tree {
	ctx, cancel := context.WithCancel(parent)

	opts := &Options{
		Protect: false,
		Create:  pipekit.NewChain[*Service](),
		Error:   pipekit.NewChain[error](),
		Wakeup:  pipekit.NewChain[*Service](),
		Panic:   pipekit.NewChain[error](),
		Report:  pipekit.NewChain[*Tree](),
	}

	for _, option := range options {
		option(opts)
	}

	tree := &Tree{}
	tree.private.context = ctx
	tree.private.cancel = cancel
	tree.private.parent = kit
	tree.private.protect = opts.Protect
	tree.handler.Report = opts.Report
	tree.handler.Create = opts.Create
	tree.handler.Error = opts.Error
	tree.handler.Wakeup = opts.Wakeup
	tree.handler.Panic = opts.Panic
	return tree
}
