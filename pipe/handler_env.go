package pipe

import (
	"github.com/vela-public/onekit/deflect"
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
	Mode    HandleType
	Parent  *lua.LState
	Pool    *LuaThreadPool
}

func (he *HandleEnv) PCall(fn *lua.LFunction, ctx *Context) error {
	cp := lua.P{
		Protect: he.Protect,
		NRet:    0,
		Fn:      fn,
	}

	co := he.Pool.Main
	if he.Mode == ReuseCo {
		co = he.Pool.Clone(he.Parent)
		defer he.Pool.Put(co)
	}

	co.Pipe = lua.NewGeneric[*Context](ctx)
	sz := len(ctx.data)
	param := make([]lua.LValue, sz)
	for i := 0; i < sz; i++ {
		item := ctx.data[i]
		param[i] = deflect.ToLValueL(co, item)
	}

	err := co.CallByParam(cp, param...)
	return err
}

func Reuse(L *lua.LState, global bool) func(*HandleEnv) {
	return func(e *HandleEnv) {
		e.Mode = ReuseCo
		if global {
			e.Pool = DefaultLuaThreadPool
		} else {
			e.Pool = NewLuaThreadPool(L)
		}
	}
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
