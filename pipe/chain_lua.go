package pipe

import "github.com/vela-public/onekit/lua"

func (c *Chain) String() string                         { return "pipe.chain" }
func (c *Chain) Type() lua.LValueType                   { return lua.LTObject }
func (c *Chain) AssertFloat64() (float64, bool)         { return float64(c.Len()), true }
func (c *Chain) AssertString() (string, bool)           { return "", false }
func (c *Chain) Hijack(fsm *lua.CallFrameFSM) bool      { return false }
func (c *Chain) AssertFunction() (*lua.LFunction, bool) { return lua.NewFunction(c.InvokeL), true }

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

func (c *Chain) ErrorL(L *lua.LState) int {
	sub := Lua(L, LState(L), Seek(1))
	if sz := len(c.private.ErrHandle); sz > 0 {
		c.private.ErrHandle = append(c.private.ErrHandle, sub.handle...)
	}
	return 0
}

func (c *Chain) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "len":
		return lua.LInt(c.Len())
	case "do":
		return lua.NewFunction(c.InvokeL)
	case "err":
		return lua.NewFunction(c.ErrorL)
	}
	return lua.LNil
}
