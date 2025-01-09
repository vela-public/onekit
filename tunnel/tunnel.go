package tunnel

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/lua"
	tun "github.com/vela-ssoc/vela-tunnel"
)

var setting = struct {
	Router *Router
	Tunnel tun.Tunneler
}{
	Router: &Router{
		cache: make(map[string]fasthttp.RequestHandler, 32),
		route: router.New(),
	},
}

func Preload(p lua.Preloader) {
	kv := lua.NewUserKV()
	kv.Set("router", lua.NewGeneric[*Router](setting.Router))
	p.Set("tunnel", kv)
}
