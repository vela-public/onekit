package taskit

import (
	"github.com/vela-public/onekit/lua"
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
	return 0
}

func (t *task) NewTypeForL(L *lua.LState) int {
	return 0
}
