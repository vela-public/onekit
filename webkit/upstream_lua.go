package webkit

import (
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/treekit"
	"reflect"
	"strings"
)

type UpstreamConfig struct {
	Name  string            `lua:"name"`
	Peers map[string]Weight `lua:"peers"`
}

type Upstream struct {
	config *UpstreamConfig
	srv    *ReverseProxy
}

func (u *Upstream) Metadata() libkit.DataKV[string, any] {
	return nil
}
func (u *Upstream) Name() string   { return u.config.Name }
func (u *Upstream) TypeOf() string { return reflect.TypeOf(u).String() }
func (u *Upstream) Close() error   { return nil }
func (u *Upstream) Start() error {
	srv, err := NewReverseProxyWith(WithBalancer(u.config.Peers))
	if err != nil {
		return err
	}
	u.srv = srv
	return nil
}

func (u *Upstream) reset() {
	if u.srv == nil {
		return
	}
	old := u.srv
	defer old.Close()
	u.config.Peers = nil
	u.srv = nil
}

func (u *Upstream) runL(L *lua.LState) int {
	if u.srv == nil {
		return 0
	}

	ctx := CheckMetadataCtx(L)
	u.srv.ServeHTTP(ctx)
	return 0
}

func (u *Upstream) OnErrorL(L *lua.LState) int {
	if u.srv == nil {
		return 0
	}

	u.srv.OnError = func(r *fasthttp.Request, res *fasthttp.Response, err error) {
		text := err.Error()
		for k, _ := range u.config.Peers {
			text = strings.ReplaceAll(text, k, "******:*")
		}

		res.SetBody([]byte(text))
	}
	return 0
}

func (u *Upstream) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "run":
		return lua.NewFunction(u.runL)
	case "vague":
		return lua.NewFunction(u.OnErrorL)
	}
	return lua.LNil
}

func (u *Upstream) Build(v *UpstreamConfig) error {
	u.config = v
	return nil
}

func (u *Upstream) Update(v *UpstreamConfig) error {
	return nil
}

func NewProxyUpstream(L *lua.LState) int {
	pro := treekit.Lazy[Upstream, UpstreamConfig](L, 1)
	pro.Build(func(conf *UpstreamConfig) *Upstream {
		return &Upstream{config: conf}
	})

	pro.Rebuild(func(conf *UpstreamConfig, u *Upstream) {
		u.config = conf
		u.reset()
	})

	pro.Start()
	L.Push(pro.Unwrap())
	return 1
}
