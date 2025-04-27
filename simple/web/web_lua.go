package web

import (
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/treekit"
)

func (hs *HttpSrv) Metadata() libkit.DataKV[string, any] {
	return nil
}

func (hs *HttpSrv) startL(L *lua.LState) int {
	treekit.Start(L, hs, L.PanicErr)
	return 0
}

func (hs *HttpSrv) RL(L *lua.LState) int {
	L.Push(hs.cnf.Router)
	return 1
}

func (hs *HttpSrv) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "r":
		return lua.NewFunction(hs.RL)
	case "start":
		return lua.NewFunction(hs.startL)
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE":
		return hs.cnf.Router.Index(L, key)
	case "not_found":
		return lua.NewFunction(hs.cnf.Router.NotFoundL)
	case "pprof":
		return lua.NewFunction(hs.cnf.Router.NewPprofL)
	}
	return lua.LNil
}
