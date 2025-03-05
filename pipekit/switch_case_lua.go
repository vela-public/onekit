package pipekit

import (
	"github.com/vela-public/onekit/lua"
)

func (c *Case[T]) push(L *lua.LState) int {
	L.Push(c)
	return 1
}

func (c *Case[T]) invokeL(L *lua.LState) int {
	happy := Lua[T](L, LState(L))
	c.Happy = happy
	return c.push(L)
}

func (c *Case[T]) breakL(L *lua.LState) int {
	c.Break = L.IsTrue(1)
	return c.push(L)
}

func (c *Case[T]) debugL(L *lua.LState) int {
	debug := LValue[*Context[T]](L, LState(L))
	c.Debug = debug
	return c.push(L)
}

func (c *Case[T]) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "invoke", "pipe", "do":
		return lua.NewFunction(c.invokeL)
	case "one":
		return lua.NewFunction(c.breakL)
	case "debug":
		return lua.NewFunction(c.debugL)
	case "raw", "text":
		return lua.LString(c.Cnd.String())
	}
	return lua.LNil
}
