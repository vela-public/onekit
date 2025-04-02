package pipe

import (
	"github.com/vela-public/onekit/lua"
)

type HandleEnv struct {
	Protect bool
	Seek    int
	Error   func(*Context, error)
	Parent  *lua.LState
}

func (he *HandleEnv) PCall(fn *lua.LFunction, ctx *Context) error {
	cp := lua.P{
		Protect: he.Protect,
		NRet:    0,
		Fn:      fn,
	}

	sz := len(ctx.data)
	co := he.Parent.Coroutine()
	defer func() {
		he.Parent.Keepalive(co)
	}()

	param := make([]lua.LValue, sz)
	for i := 0; i < sz; i++ {
		item := ctx.data[i]
		param[i] = lua.ReflectTo(item)
	}

	err := co.CallByParam(cp, param...)
	return err
}

func Protect(b bool) func(*HandleEnv) {
	return func(e *HandleEnv) {
		e.Protect = b
	}
}

func Clone(e *HandleEnv) func(env *HandleEnv) {
	return func(env *HandleEnv) {
		env.Protect = e.Protect
		env.Seek = e.Seek
		env.Error = e.Error
		env.Parent = e.Parent
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
