package treekit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
)

func CheckName(v string) error {
	if len(v) < 2 {
		return fmt.Errorf("too short name got:%s", v)
	}

	if !IsChar(v[0]) {
		return fmt.Errorf("first char must be a-z or A-Z got:%v", string(v[0]))
	}

	n := len(v)
	for i := 1; i < n; i++ {
		ch := v[i]
		switch {
		case IsChar(ch), IsInt(ch):
			continue
		case ch == '_':
			continue
		case ch == '-':
			continue
		case ch == '/':
			continue
		case ch == '|':
			continue
		default:
			return fmt.Errorf("not allowed char %v", string(ch))

		}
	}
	return nil

}

func Check[T any](L *lua.LState, pro *Process) (t T) {
	if pro.Nil() {
		L.RaiseError("not found processes data")
		return
	}

	dat, ok := pro.data.(T)
	if !ok {
		L.RaiseError("mismatch processes type must:%T got:%T", t, pro.data)
		return t
	}

	return dat
}

func Create(L *lua.LState, name string, typeof string) *Process {
	if err := CheckName(name); err != nil {
		L.RaiseError("%v", err)
		return nil
	}
	exdata := L.Exdata()
	switch dat := exdata.(type) {
	case *MicroService:
		return dat.Create(L, name, typeof)
	case *Task:
		return dat.Create(L, name, typeof)
	default:
		L.RaiseError("lua.exdata must *MicroService or *TaskTree got:%T", exdata)
		return nil
	}

}

// LazyCreate 不检查name是否合法
func LazyCreate(L *lua.LState, name string, typeof string) *Process {
	exdata := L.Exdata()
	switch dat := exdata.(type) {
	case *MicroService:
		return dat.Create(L, name, typeof)
	case *Task:
		return dat.Create(L, name, typeof)
	default:
		L.RaiseError("lua.exdata must *MicroService or *TaskTree got:%T", exdata)
		return nil
	}
}

func Start(L *lua.LState, v ProcessType, x func(e error)) {
	exdata := L.Exdata()
	switch dat := exdata.(type) {
	case *MicroService:
		dat.Startup(v, x)
	case *Task:
		dat.Startup(v, x)
	default:
		x(fmt.Errorf("lua.exdata must *MicroService or *TaskTree got:%T", exdata))
	}
}
