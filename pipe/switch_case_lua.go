package pipe

import (
	"github.com/vela-public/onekit/lua"
)

func (c *Case) push(L *lua.LState) int {
	L.Push(c)
	return 1
}

func (c *Case) invokeL(L *lua.LState) int {
	happy := Lua(L, LState(L), Seek(1))
	c.Happy = happy
	return c.push(L)
}

func (c *Case) breakL(L *lua.LState) int {
	c.Break = L.IsTrue(1)
	return c.push(L)
}

func (c *Case) debugL(L *lua.LState) int {
	debug := Lua(L, LState(L))
	c.Debug = debug
	return c.push(L)
}

func (c *Case) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "invoke", "pipe", "do":
		return lua.NewFunction(c.invokeL)
	case "one":
		return lua.NewFunction(c.breakL)
	case "debug":
		return lua.NewFunction(c.debugL)

	}
	return lua.LNil
}
