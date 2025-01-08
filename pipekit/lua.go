package pipekit

import "github.com/vela-public/onekit/lua"

func Lua[T any](L *lua.LState, options ...func(*HandleEnv)) *Chain[T] {
	env := NewEnv(options...)
	c := &Chain[T]{}

	if env.Seek == 0 {
		env.Seek = 1
	}

	n := L.GetTop()
	if n-env.Seek < 0 {
		return c
	}

	for idx := env.Seek; idx <= n; idx++ {
		c.handler(L.Get(idx), env)
	}
	return c
}

func LValue[T any](v lua.LValue, options ...func(*HandleEnv)) *Chain[T] {
	env := NewEnv(options...)
	c := &Chain[T]{}
	c.handler(v, env)
	return c
}

func ChainL(L *lua.LState) int {
	v := Lua[lua.LValue](L, LState(L))
	L.Push(v)
	return 1
}

func SwitchL(L *lua.LState) int {
	v := lua.NewGeneric[*Switch[lua.LValue]](NewSwitch[lua.LValue]())
	L.Push(v)
	return 1
}

func Preload(v lua.Preloader) {
	v.Set("pipekit", lua.NewExport("lua.pipe.export", lua.WithFunc(ChainL)))
	v.Set("switch2", lua.NewExport("lua.switch.export", lua.WithFunc(SwitchL)))
}
