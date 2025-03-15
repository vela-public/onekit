package treekit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"strings"
)

func number(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

func alphabet(ch byte) bool {
	if ch >= 'a' && ch <= 'z' {
		return true
	}

	if ch >= 'A' && ch <= 'Z' {
		return true
	}

	return false
}

func Name(v string) error {
	if len(v) < 2 {
		return fmt.Errorf("too short name got:%s", v)
	}

	if !alphabet(v[0]) {
		return fmt.Errorf("first char must be a-z or A-Z got:%v", string(v[0]))
	}

	if strings.HasPrefix(v, "GET /") ||
		strings.HasPrefix(v, "POST /") ||
		strings.HasPrefix(v, "PUT /") ||
		strings.HasPrefix(v, "DELETE /") ||
		strings.HasPrefix(v, "HEAD /") ||
		strings.HasPrefix(v, "PATCH /") ||
		strings.HasPrefix(v, "OPTIONS /") {

		if offset := strings.IndexFunc(v, func(r rune) bool {
			return r == '/'
		}); offset != -1 {
			v = v[offset:]
		}
	}

	sz := len(v)
	for i := 1; i < sz; i++ {
		ch := v[i]
		switch {
		case alphabet(ch), number(ch):
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
