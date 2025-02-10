package pipekit

import (
	"github.com/vela-public/onekit/lua"
)

type LuaPool interface {
	Coroutine() *lua.LState
	Clone(*lua.LState) *lua.LState
	Put(*lua.LState)
}

type HandleEnv struct {
	Protect bool
	Seek    int
	Parent  *lua.LState
}

func Protect(b bool) func(*HandleEnv) {
	return func(e *HandleEnv) {
		e.Protect = b
	}
}

func LState(co *lua.LState) func(*HandleEnv) {
	return func(e *HandleEnv) {
		e.Parent = co
	}
}

func Seek(n int) func(*HandleEnv) {
	return func(e *HandleEnv) {
		e.Seek = n
	}
}

func NewEnv(opts ...func(*HandleEnv)) *HandleEnv {
	env := &HandleEnv{}

	for _, opt := range opts {
		opt(env)
	}
	return env
}
