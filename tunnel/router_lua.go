package tunnel

import "github.com/vela-public/onekit/lua"

func (trr *TRouter) index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "GET", "POST", "PUT", "PATCH":
		return trr.newHandleL(L, key)
	case "client":
		cli := &Client{table: trr}
		return cli
	}

	return lua.LNil
}
