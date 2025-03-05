package tunnel

import "github.com/vela-public/onekit/lua"

func (rr *Router) HandleL(L *lua.LState, method string) lua.LValue {
	return NewHandleL(L, rr, method)
}

func (rr *Router) index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE", "PATCH":
		return rr.HandleL(L, key)
	case "client":
		cli := &Client{table: rr}
		return cli
	}

	return lua.LNil
}
