package webkit

import (
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/treekit"
	"reflect"
	"strings"
)

type Upstream struct {
	name  string
	peers map[string]Weight
	srv   *ReverseProxy
}

func (u *Upstream) Metadata() libkit.DataKV[string, any] {
	return nil
}
func (u *Upstream) Name() string   { return u.name }
func (u *Upstream) TypeOf() string { return reflect.TypeOf(u).String() }
func (u *Upstream) Close() error   { return nil }
func (u *Upstream) Start() error {
	srv, err := NewReverseProxyWith(WithBalancer(u.peers))
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
	u.peers = nil
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
		for k, _ := range u.peers {
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

func NewProxyUpstream(L *lua.LState) int {
	name := L.CheckString(1)
	ups := &Upstream{name: name}
	srv := treekit.Create(L, name, ups.TypeOf())
	if srv.Nil() {
		srv.Set(ups)
	} else {
		srv.Call(func(dat treekit.ProcessType) {
			ups = dat.(*Upstream)
			ups.reset()
		})
		srv.Reload()
	}

	tab := L.CheckTable(2)
	tab.Range(func(peer string, w lua.LValue) {
		peers := make(map[string]Weight)
		peers[peer] = Weight(lua.CheckInt(L, w))
		ups.peers = peers
	})

	treekit.Start(L, ups, func(e error) {
		L.RaiseError("%v", e)
	})
	L.Push(srv)
	return 1
}
