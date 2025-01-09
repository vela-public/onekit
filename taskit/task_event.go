package taskit

import (
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
	return 0
}

func (t *task) NewTaskDebugL(L *lua.LState) int {
	line := L.Where(1)
	text := "[" + line[:len(line)-1] + "]" + luakit.Format(L, 0)
	event := layer.Debug(text)
	event.FromCode = t.Key()
	event.Log()
	if t.setting.Debug {
		event.Put()
	}
	return 0
}

func (t *task) NewTypeForL(L *lua.LState) int {
	return 0
}
