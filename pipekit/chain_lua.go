package pipekit

import (
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
)

func (c *Chain[T]) InvokeL(L *lua.LState) int {
	n := L.GetTop()
	args := make([]T, n)
	for i := 1; i <= n; i++ {
		args[i-1] = luakit.Check[T](L, L.Get(i))
	}
	v := c.Invoke(args...)
	L.Push(v)
	return 1
}

func (c *Chain[T]) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(c.InvokeL), true
}
