package treekit

import (
	"context"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/pipe"
	"strings"
	"sync"
)

type MsTree struct {
	cache struct {
		mutex sync.RWMutex
		data  []*MicroService
	}

	handler struct {
		Report *pipe.Chain
		Create *pipe.Chain
		Error  *pipe.Chain
		Wakeup *pipe.Chain
		Panic  *pipe.Chain
	}

	private struct {
		protect bool
		context context.Context
		cancel  context.CancelFunc
		luakit  *luakit.Kit
		error   error
	}
}

func (mt *MsTree) Protect() bool {
	return mt.private.protect
}

func (mt *MsTree) Length() int {
	return len(mt.cache.data)
}

func (mt *MsTree) BeLinked(name string) bool {
	for _, ms := range mt.cache.data {
		if libkit.In[string](ms.processes.Link, name) {
			return true
		}
	}
	return false
}

func (mt *MsTree) Lookup(L *lua.LState) int {
	xpath := L.CheckString(1)
	dst := strings.Split(xpath, ".")
	if len(dst) != 2 {
		L.RaiseError("import name is empty")
		return 0
	}

	key := dst[0]
	name := dst[1]

	tas, ok := mt.find(key)
	if !ok {
		L.RaiseError("not found %s", key)
		return 0
	}

	if tas.Key() == L.Name() {
		L.RaiseError("loop call %s", L.Name())
		return 0
	}

	//提前唤醒
	if tas.has(Register) {
		_ = tas.wakeup()
	}

	pro, ok := tas.have(name)
	if !ok {
		L.RaiseError("not found %s", key)
		return 0
	}

	L.Push(pro)
	return 1
}

func (mt *MsTree) LuaKit() *luakit.Kit {
	return mt.private.luakit.Clone()
}

func (mt *MsTree) Context() context.Context {
	return mt.private.context
}

func (mt *MsTree) push(tas *MicroService) {
	mt.cache.mutex.Lock()
	defer mt.cache.mutex.Unlock()
	mt.cache.data = append(mt.cache.data, tas)
}

func (mt *MsTree) find(key string) (*MicroService, bool) {
	mt.cache.mutex.RLock()
	defer mt.cache.mutex.RUnlock()

	sz := len(mt.cache.data)
	if sz == 0 {
		return nil, false
	}

	for i := 0; i < sz; i++ {
		tas := mt.cache.data[i]
		if tas.config.Key == key {
			return tas, true
		}
	}

	return nil, false
}

func (mt *MsTree) RemoveByID(ids []int64) {
	mt.cache.mutex.Lock()
	defer mt.cache.mutex.Unlock()
	var mss []*MicroService

	for i, ms := range mt.cache.data {
		if !libkit.In[int64](ids, ms.ID()) {
			mss = append(mss, mt.cache.data[i])
			continue
		}
		ms.Close()
	}
	mt.cache.data = mss
}

func NewMicoSrvTree(parent context.Context, kit *luakit.Kit, option *MicroServiceOption) *MsTree {
	ctx, cancel := context.WithCancel(parent)
	tree := &MsTree{}
	tree.private.context = ctx
	tree.private.cancel = cancel
	tree.private.luakit = kit
	tree.private.protect = option.protect
	tree.handler.Report = option.report
	tree.handler.Create = option.create
	tree.handler.Error = option.error
	tree.handler.Wakeup = option.wakeup
	tree.handler.Panic = option.panic
	return tree
}
