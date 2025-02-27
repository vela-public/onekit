package pipe

import (
	"github.com/vela-public/onekit/lua"
)

type HandleEnv struct {
	Protect bool
	Seek    int
	Mode    HandleType
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
