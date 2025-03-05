package treekit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
)

func (pro *Process) String() string                         { return fmt.Sprintf("%p", pro) }
func (pro *Process) Type() lua.LValueType                   { return lua.LTObject }
func (pro *Process) AssertFloat64() (float64, bool)         { return 0, false }
func (pro *Process) AssertString() (string, bool)           { return "", false }
func (pro *Process) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (pro *Process) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (pro *Process) Text() string {
	if pro.data == nil {
		return fmt.Sprintf("task.processes<%p>", pro)
	}

	str, ok := pro.data.(fmt.Stringer)
	if ok {
		return str.String()
	}

	return fmt.Sprintf("task.processes<%p>", pro)
}

func (pro *Process) Private(L *lua.LState) {
	data := L.Exdata()
	switch dat := data.(type) {
	case *MicroService:
		if dat.Key() != pro.from {
			L.RaiseError("%s processes.from=%s with %s not allow", dat.Key(), pro.from, pro.name)
			return
		}
		pro.private = true
	case *Task:
		pro.private = true
	default:
		L.RaiseError("not found service or task exdata")
		return
	}
}

func (pro *Process) Index(L *lua.LState, key string) lua.LValue {
	if pro.data == nil {
		return lua.LNil
	}

	it, ok := pro.data.(lua.IndexType)
	if ok {
		return it.Index(L, key)
	}
	return lua.LNil
}

func (pro *Process) Meta(L *lua.LState, key lua.LValue) lua.LValue {
	if pro.data == nil {
		return lua.LNil
	}

	it, ok := pro.data.(lua.MetaType)
	if ok {
		return it.Meta(L, key)
	}
	return lua.LNil
}

func (pro *Process) NewMeta(L *lua.LState, key lua.LValue, val lua.LValue) {
	if pro.data == nil {
		return
	}

	it, ok := pro.data.(lua.NewMetaType)
	if ok {
		it.NewMeta(L, key, val)
		return
	}
}

func (pro *Process) MetaTable(L *lua.LState, key string) lua.LValue {
	if pro.data == nil {
		return lua.LNil
	}
	it, ok := pro.data.(lua.MetaTableType)
	if ok {
		return it.MetaTable(L, key)
	}
	return lua.LNil
}

func (pro *Process) NewIndex(L *lua.LState, key string, val lua.LValue) {
	if pro.data == nil {
		return
	}

	it, ok := pro.data.(lua.NewIndexType)
	if ok {
		it.NewIndex(L, key, val)
		return
	}
}
