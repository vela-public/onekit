package tunnel

import (
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/webkit"
)

type Client struct {
	table *TRouter
}

func (c Client) String() string                         { return "tunnel.router.client" }
func (c Client) Type() lua.LValueType                   { return lua.LTObject }
func (c Client) AssertFloat64() (float64, bool)         { return 0, false }
func (c Client) AssertString() (string, bool)           { return "", false }
func (c Client) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c Client) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (c Client) Exec(method string, L *lua.LState) int {
	req := L.CheckString(1)   //req url
	data := L.Get(2).String() //req body
	_, _ = c.table.Exec(method, req, data)
	r, err := c.table.Exec(method, req, data)
	rsp := webkit.NewResponseL(r, err)
	L.Push(rsp)
	return 0
}

func (c Client) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "GET":
		return L.NewFunction(func(co *lua.LState) int {
			return c.Exec("GET", co)
		})

	case "POST":
		return L.NewFunction(func(co *lua.LState) int {
			return c.Exec("POST", co)
		})
	}

	L.RaiseError("not found %s with router client", key)
	return lua.LNil
}
