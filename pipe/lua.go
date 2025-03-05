package pipe

import (
	"github.com/vela-public/onekit/lua"
)

/*
	1
	pipe.seek(1) func(p1 , p2)
	pipe.Lua(L , pipe.LState(L) , pipe.Seek(1)))
*/

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
	v := lua.NewGeneric[*Chain](Lua(L, LState(L)))
	L.Push(v)
	return 1
}

func SwitchL(L *lua.LState) int {
	v := NewSwitch()
	L.Push(v)
	return 1
}

func Preload(v lua.Preloader) {
	kv := lua.NewUserKV()
	kv.Set("switch", lua.NewFunction(SwitchL))
	v.SetGlobal("pipe", lua.NewExport("lua.pipe.export", lua.WithFunc(ChainL), lua.WithTable(kv)))
}
