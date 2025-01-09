package tunnel

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-ssoc/vela-common-mba/definition"
	tun "github.com/vela-ssoc/vela-tunnel"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Transport struct {
	Tunnel  tun.Tunneler
	private struct {
		Error   error
		MHide   definition.MHide
		Context context.Context
		Router  *Router
	}
}

func (tp *Transport) ID() string {
	if tp.Tunnel == nil {
		return ""
	}
	return strconv.FormatInt(tp.Tunnel.Issue().ID, 10)
}

func (tp *Transport) Broker() (net.IP, int) {
	if tp.Tunnel == nil {
		return nil, 0
	}

	addr, ok := tp.Tunnel.RemoteAddr().(*net.TCPAddr)
	if ok {
		return addr.IP, addr.Port
	}
	return nil, 0
}

func (tp *Transport) R() layer.RouterType {
	return tp.private.Router
}

func (tp *Transport) Node() string {
	if tp.Tunnel == nil {
		return ""
	}
	return tp.Tunnel.NodeName()
}

func (tp *Transport) Tags() []string {
	return tp.private.MHide.Tags
}

func (tp *Transport) Doer(prefix string) (layer.Doer, error) {
	if tp.Tunnel == nil {
		return nil, NoTunnelE
	}
	dr := tp.Tunnel.Doer(prefix)
	if dr == nil {
		return nil, fmt.Errorf("tunnel doer %s failed", prefix)
	}
	return dr, nil
}

func (tp *Transport) Oneway(path string, reader io.Reader, header http.Header) error {
	if tp.Tunnel == nil {
		return NoTunnelE
	}
	ctx, cancel := context.WithTimeout(tp.private.Context, 10*time.Second)
	defer cancel()

	return tp.Tunnel.Oneway(ctx, path, reader, header)
}

func (tp *Transport) Fetch(path string, reader io.Reader, header http.Header) (*http.Response, error) {
	if tp.Tunnel == nil {
		return nil, NoTunnelE
	}
	ctx, cancel := context.WithTimeout(tp.private.Context, 10*time.Second)
	defer cancel()

	return tp.Tunnel.Fetch(ctx, path, reader, header)
}

func (tp *Transport) JSON(path string, data interface{}, result interface{}) error {
	if tp.Tunnel == nil {
		return NoTunnelE
	}
	ctx, cancel := context.WithTimeout(tp.private.Context, 10*time.Second)
	defer cancel()

	return tp.Tunnel.JSON(ctx, path, data, result)
}

func (tp *Transport) Push(path string, data interface{}) error {
	if tp.Tunnel == nil {
		return NoTunnelE
	}
	ctx, cancel := context.WithTimeout(tp.private.Context, 10*time.Second)
	defer cancel()

	return tp.Tunnel.OnewayJSON(ctx, path, data)
}

func (tp *Transport) Stream(ctx context.Context, s string, header http.Header) (*websocket.Conn, error) {
	//TODO implement me
	panic("implement me")
}

func (tp *Transport) Attachment(name string) (layer.Attachment, error) {
	//TODO implement me
	panic("implement me")
}

func (tp *Transport) Preload(p lua.Preloader) {
	kv := lua.NewUserKV()
	kv.Set("router", lua.NewGeneric[*Router](tp.private.Router))
	p.Set("tunnel", kv)
}

func NewTransport(ctx context.Context) *Transport {
	t := &Transport{}
	t.private.Router = NewRouter()
	t.private.Context = ctx
	return t
}

func (tp *Transport) Devel(ctx context.Context, vip, edition, host string, options ...tun.Option) {
	hide := definition.MHide{
		Addrs:      []string{vip},
		Servername: host,
		Semver:     edition,
	}

	tp.private.MHide = hide

	tnl, err := tun.Dial(ctx, hide, tp.private.Router.h2s(), options...)
	if err != nil {
		tp.private.Error = err
		return
	}
	tp.Tunnel = tnl
}

func Worker(ctx context.Context) *Transport {
	return nil
}
