package taskit

import (
	"github.com/vela-public/onekit/event"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
)

func (t *task) NewTaskEventL(L *lua.LState) int {
	return 0
}

func (t *task) NewTaskTraceL(L *lua.LState) int {
	return 0
}

func (t *task) NewTaskErrorL(L *lua.LState) int {
	xEnv := layer.LazyEnv()
	line := L.Where(1)
	text := "[" + line[:len(line)-1] + "]" + luakit.Format(L, 0)
	ev := event.Error(xEnv, text)
	ev.FromCode = t.Key()
	ev.Error(xEnv.Logger())
	if t.setting.Debug {
		ev.Put(xEnv.Transport())
	}
	return 0
}

func (t *task) NewTaskDebugL(L *lua.LState) int {
	xEnv := layer.LazyEnv()
	line := L.Where(1)
	text := "[" + line[:len(line)-1] + "]" + luakit.Format(L, 0)
	ev := event.Debug(xEnv, text)
	ev.FromCode = t.Key()
	ev.Debug(xEnv.Logger())
	if t.setting.Debug {
		ev.Put(xEnv.Transport())
	}
	return 0
}

func (t *task) NewTypeForL(L *lua.LState) int {
	xEnv := layer.LazyEnv()
	ev := event.Debug(xEnv, luakit.TypeFor(L))
	ev.FromCode = t.Key()
	ev.Debug(xEnv.Logger())
	if t.setting.Debug {
		ev.Put(xEnv.Transport())
	}
	return 0
}
