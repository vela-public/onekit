package tunnel

import "github.com/vela-public/onekit/lua"

func (rr *Router) index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "GET", "POST", "PUT", "PATCH":
		return rr.newHandleL(L, key)
	case "client":
		cli := &Client{table: rr}
		return cli
	}

	return lua.LNil
}
