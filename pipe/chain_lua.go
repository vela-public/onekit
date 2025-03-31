package pipe

import "github.com/vela-public/onekit/lua"

func (c *Chain) InvokeL(L *lua.LState) int {
	n := L.GetTop()
	args := make([]any, n)
	for i := 1; i <= n; i++ {
		args[i-1] = L.Get(i)
	}

	v := c.Invoke(args...)
	L.Push(v)
	return 1
}

func (c *Chain) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(c.InvokeL), true
}
