package tunnel

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/lua"
	tun "github.com/vela-ssoc/vela-tunnel"
)

var setting = struct {
	Router *TRouter
	Tunnel tun.Tunneler
}{
	Router: &TRouter{
		cache: make(map[string]fasthttp.RequestHandler, 32),
		route: router.New(),
	},
}

func Preload(p lua.Preloader) {
	kv := lua.NewUserKV()
	kv.Set("router", lua.NewGeneric[*TRouter](setting.Router))
	p.Set("tunnel", kv)
}
