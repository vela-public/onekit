package pipe

import "github.com/vela-public/onekit/lua"

func Lua(L *lua.LState, options ...func(*HandleEnv)) *Chain {
	env := NewEnv(options...)
	c := &Chain{}

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

func LValue(v lua.LValue, options ...func(*HandleEnv)) *Chain {
	env := NewEnv(options...)
	c := &Chain{}
	c.handler(v, env)
	return c
}

func ChainL(L *lua.LState) int {
	v := lua.NewGeneric[*Chain](Lua(L, LState(L), Reuse(L, true)))
	L.Push(v)
	return 1
}

func SwitchL(L *lua.LState) int {
	v := lua.NewGeneric[*Switch](NewSwitch())
	L.Push(v)
	return 1
}

func Preload(v lua.Preloader) {
	v.Set("pipe", lua.NewExport("lua.pipe.export", lua.WithFunc(ChainL)))
	v.Set("switch", lua.NewExport("lua.switch.export", lua.WithFunc(SwitchL)))
}
