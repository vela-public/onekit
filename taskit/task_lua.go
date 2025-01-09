package taskit

import (
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"strings"
)

func (t *task) debugL(L *lua.LState) int {
	t.setting.Debug = lua.IsTrue(L.Get(1))
	return 0
}

func (t *task) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "debug":
		return lua.NewFunction(t.debugL)

	}
	return lua.LNil
}

func (t *task) startL(L *lua.LState) int {
	srv := luakit.Check[*Service](L, L.Get(1))
	Start(L, srv.data, func(e error) {
		L.RaiseError("%v", e)
	})
	return 0
}

func (t *task) privateL(L *lua.LState) int {
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		srv := luakit.Check[*Service](L, L.Get(i))
		srv.Private(L)
	}
	return 0
}

func (t *task) disableL(L *lua.LState) int {
	t.put(Disable)
	L.RaiseError("disable")
	return 0
}

func (t *task) LinkL(L *lua.LState) int {
	dst := strings.Split(L.CheckString(1), ".")
	if len(dst) != 2 {
		L.RaiseError("import name is empty")
		return 0
	}

	key := dst[0]
	name := dst[1]

	tas, ok := t.root.find(key)
	if !ok {
		L.RaiseError("not found %s", key)
		return 0
	}

	if tas.Key() == t.Key() {
		L.RaiseError("loop call %s", t.Key())
		return 0
	}
	t.link(key)

	//提前唤醒
	if tas.has(Register) {
		_ = tas.wakeup()
	}

	srv, ok := tas.have(name)
	if !ok {
		L.RaiseError("not found %s", key)
		return 0
	}

	L.Push(srv)
	return 1
}

func (t *task) Preload(kit *luakit.Kit) { //luakit.error() luakit.trace() luakit.T()
	kit.Set("start", lua.NewFunction(t.startL))
	kit.Set("private", lua.NewFunction(t.privateL))
	kit.Set("disable", lua.NewFunction(t.disableL))
	kit.Set("event", lua.NewFunction(t.NewTaskEventL))
	kit.Set("trace", lua.NewFunction(t.NewTaskTraceL))
	kit.Set("error", lua.NewFunction(t.NewTaskErrorL))
	kit.Set("debug", lua.NewFunction(t.NewTaskDebugL))
	kit.Set("T", lua.NewFunction(t.NewTypeForL))
	kit.SetGlobal("this", lua.NewGeneric[*task](t))
	kit.SetGlobal("import", lua.NewExport("lua.taskit.export", lua.WithFunc(t.LinkL)))
}
