package tunnel

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-ssoc/vela-common-mba/definition"
	tun "github.com/vela-ssoc/vela-tunnel"
)

func Push(ctx context.Context, path string, data interface{}) error {
	return setting.Tunnel.OnewayJSON(ctx, path, data)
}

func Connect(ctx context.Context, peer, edit, host string, opts ...tun.Option) error {
	tnl, err := tun.Dial(ctx, definition.MHide{
		Addrs:      []string{peer},
		Servername: host,
		Semver:     edit,
	}, setting.Router.h2s(), opts...)

	if err != nil {
		return fmt.Errorf("tunnel v2 connect fail %v", err)
	}

	setting.Tunnel = tnl
	return nil
}

func R() layer.RouterType {
	return setting.Router
}

func Handle(method, path string, handle fasthttp.RequestHandler) error {
	return setting.Router.Handle(method, path, handle)
}

func PUT(path string, handle fasthttp.RequestHandler) error {
	return setting.Router.Handle("PUT", path, handle)
}

func GET(path string, handle fasthttp.RequestHandler) error {
	return setting.Router.Handle("GET", path, handle)
}

func POST(path string, handle fasthttp.RequestHandler) error {
	return setting.Router.Handle("POST", path, handle)
}

func DELETE(path string, handle fasthttp.RequestHandler) error {
	return setting.Router.Handle("DELETE", path, handle)
}
