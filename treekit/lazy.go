package treekit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
)

type LazyTree struct{}

func Lazy() *LazyTree {
	return &LazyTree{}
}

func (l *LazyTree) Int(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

func (l *LazyTree) Alphabet(ch byte) bool {
	if ch >= 'a' && ch <= 'z' {
		return true
	}

	if ch >= 'A' && ch <= 'Z' {
		return true
	}

	return false
}

func (l *LazyTree) Name(v string) error {
	if len(v) < 2 {
		return fmt.Errorf("too short name got:%s", v)
	}

	if !l.Alphabet(v[0]) {
		return fmt.Errorf("first char must be a-z or A-Z got:%v", string(v[0]))
	}

	n := len(v)
	for i := 1; i < n; i++ {
		ch := v[i]
		switch {
		case l.Alphabet(ch), l.Int(ch):
			continue
		case ch == '_':
			continue
		case ch == '-':
			continue
		case ch == '/':
			continue
		default:
			return fmt.Errorf("not allowed char %v", string(ch))

		}
	}
	return nil
}

func (l *LazyTree) Create(L *lua.LState, name string, typeof string) *Process {
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
