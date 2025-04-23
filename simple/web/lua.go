package web

import (
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/treekit"
)

func NewSimpleHttpL(L *lua.LState) int {
	//todo config

	pro := treekit.Lazy[HttpSrv, Config](L, 1)
	pro.Build(func(cnf *Config) *HttpSrv {
		cnf.Must()
		return NewSrv(cnf)
	})

	pro.Rebuild(func(cnf *Config, hs *HttpSrv) {
		cnf.Must()
		hs.cnf = cnf
	})

	L.Push(pro.Unwrap())
	return 1
}

/*
	local route = vela.http_route("")

	local srv = vela.simple_http_srv{
		name = "test",
        bind = "127.0.0.1:8899",
	}

	srv.before(function(ctx)
		local conn = ctx()
		conn.say()
        --ctx.say("hello")
		return 0
	end)

	srv.GET("/uri",  app1)
	srv.POST("/uri", app2)
	srv.GET("/api",  app1)
	srv.POST("/detect", app2)
	srv.start()

*/

func Preload(p lua.Preloader) {
	kv := lua.NewUserKV()
	p.Set("simple_http", lua.NewExport("lua.http_srv", lua.WithFunc(NewSimpleHttpL), lua.WithTable(kv)))
}
