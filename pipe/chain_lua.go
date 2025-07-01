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

	c.Invokes(args)
	return 0
}

func (c *Chain) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "len":
		return lua.LInt(c.Len())
	case "do":
		return lua.NewFunction(c.InvokeL)
	}
	return lua.LNil
}
