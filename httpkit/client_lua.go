package httpkit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipe"
	"strings"
)

func (c *Client) String() string                         { return fmt.Sprintf("http.client %p", c) }
func (c *Client) Type() lua.LValueType                   { return lua.LTObject }
func (c *Client) AssertFloat64() (float64, bool)         { return 0, false }
func (c *Client) AssertString() (string, bool)           { return "", false }
func (c *Client) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c *Client) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (c *Client) H(L *lua.LState) int {
	data := L.CheckString(1)
	kv := strings.Split(data, ":")

	if len(kv) == 2 {
		c.SetHeader(kv[0], kv[1])
	}
	L.Push(c)
	return 1
}

func (c *Client) afterL(L *lua.LState) int {
	chain := pipe.Lua(L, pipe.LState(L))
	if chain.Len() == 0 {
		return 0
	}
	c.OnAfterResponse(func(_ *Client, response *Response) error {
		return chain.InvokeGo(response).UnwrapErr()
	})
	return 0
}

func (c *Client) beforeL(L *lua.LState) int {
	chain := pipe.Lua(L, pipe.LState(L), pipe.Protect(true))
	if chain.Len() == 0 {
		return 0
	}
	c.OnBeforeRequest(func(_ *Client, request *Request) error {
		return chain.InvokeGo(request).UnwrapErr()
	})
	return 0
}

func (c *Client) authL(L *lua.LState) int {
	c.SetAuthToken(L.CheckString(1))
	return 0
}

func (c *Client) proxyL(L *lua.LState) int {
	c.SetProxy(L.CheckString(1))
	return 0
}

func (c *Client) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "R":
		return c.R()
	case "after":
		return lua.NewFunction(c.afterL)
	case "before":
		return lua.NewFunction(c.beforeL)
	case "auth":
		return lua.NewFunction(c.authL)
	case "proxy":
		return lua.NewFunction(c.proxyL)

	case "GET", "POST", "PUT", "HEAD", "OPTIONS", "PATCH", "DELETE", "TRACE":
		r := c.R()
		r.Method = key

		return L.NewFunction(r.exec)

	case "H":
		return L.NewFunction(c.H)

	default:

	}

	return lua.LNil
}
