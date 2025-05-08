package treekit

import (
	"context"
	"fmt"
	"github.com/vela-public/onekit/lua"
)

type Env struct {
	ctx  context.Context
	lua  *lua.LState
	err  func(error)
	data any
}

func (env *Env) Context() context.Context {
	if env.ctx == nil {
		return env.lua.Context()
	}
	return env.ctx
}

func (env *Env) ExData() any {
	return env.data
}

func (env *Env) Error(err error) {
	if env.err != nil {
		env.err(err)
	}
}

func (env *Env) Errorf(format string, args ...interface{}) {
	env.Error(fmt.Errorf(format, args...))
}

func (env *Env) LState() *lua.LState {
	return env.lua
}

type TreeEnvFunc func(*Env)

func Context(ctx context.Context) TreeEnvFunc {
	return func(te *Env) {
		te.ctx = ctx
	}
}

func LState(co *lua.LState) TreeEnvFunc {
	return func(te *Env) {
		te.lua = co
	}
}

func Exdata(data any) TreeEnvFunc {
	return func(te *Env) {
		te.data = data
	}
}

func ErrHandler(fn func(error)) TreeEnvFunc {
	return func(te *Env) {
		te.err = fn
	}
}
