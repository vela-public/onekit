package treekit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
)

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
	if err := Lazy().Name(name); err != nil {
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
