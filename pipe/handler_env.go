package pipe

import (
	"github.com/vela-public/onekit/lua"
)

type HandleEnv struct {
	Protect bool
	Seek    int
	Error   func(*Catalog, error)
	Parent  *lua.LState
}

func (he *HandleEnv) PCall(fn *lua.LFunction, ctx *Catalog) error {
	cp := lua.P{
		Protect: he.Protect,
		NRet:    0,
		Fn:      fn,
	}

	sz := len(ctx.data)
	co := he.Parent.Coroutine()

	var err error
	switch sz {
	case 0:
		err = co.CallByParam(cp)
	case 1:
		err = co.CallByParam(cp, lua.ReflectTo(ctx.data[0]))
	case 2:
		err = co.CallByParam(cp, lua.ReflectTo(ctx.data[0]), lua.ReflectTo(ctx.data[1]))
	default:
		param := make([]lua.LValue, sz)
		for i := 0; i < sz; i++ {
			item := ctx.data[i]
			param[i] = lua.ReflectTo(item)
		}
		err = co.CallByParam(cp, param...)
	}

	if err == nil {
		he.Parent.Keepalive(co)
	}

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
	env := &HandleEnv{
		Protect: true,
	}

	for _, opt := range opts {
		opt(env)
	}
	return env
}
