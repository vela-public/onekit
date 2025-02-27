package tunnel

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipekit"
	"github.com/vela-public/onekit/treekit"
	"github.com/vela-public/onekit/webkit"
	"path/filepath"
	"reflect"
	"strings"
)

var typeof = reflect.TypeOf((*LHandle)(nil)).String()

type config struct {
	uri     string
	method  string
	recycle bool
}

type LHandle struct {
	cfg     config
	trr     *Router
	co      *lua.LState
	webctx  *webkit.HttpContext
	handle  *pipekit.Chain[*webkit.WebContext]
	private struct {
		Count   int64 //请求次数
		Failed  int64 //失败次数
		Succeed int64 //成功次数
	}
}

func (lh *LHandle) Name() string {
	return fmt.Sprintf("%s %s", lh.cfg.method, lh.cfg.uri)
}

func (lh *LHandle) TypeOf() string {
	return typeof
}

func (lh *LHandle) Metadata() libkit.DataKV[string, any] {
	mt := libkit.NewDataKV[string, any]()
	mt.Set("count", lh.private.Count)
	mt.Set("failed", lh.private.Failed)
	mt.Set("succeed", lh.private.Succeed)
	return *mt
}

func (lh *LHandle) HandleFunc(ctx *fasthttp.RequestCtx) {
	webctx := &webkit.WebContext{Request: ctx}
	lh.handle.Invoke(webctx)
}

func (lh *LHandle) Start() error {
	if lh.handle == nil {
		return fmt.Errorf("router start fail not found handle")
	}
	return lh.trr.Handle(lh.cfg.method, lh.cfg.uri, lh.HandleFunc)
}

func (lh *LHandle) Close() error {
	if lh.cfg.recycle {
		return nil
	}

	lh.cfg.recycle = true
	lh.trr.Undo(lh.cfg.method, lh.cfg.uri)
	return nil
}

func (lh *LHandle) SetURI(path string) {
	var uri string
	if !strings.HasPrefix(path, "/api/v1/arr/lua/") {
		uri = filepath.Join("/api/v1/arr/lua/", lh.co.Name(), path)
	} else {
		uri = filepath.Join("/", lh.co.Name(), path)
	}
	lh.cfg.uri = filepath.ToSlash(uri)
}

func (lh *LHandle) NewSRV(L *lua.LState) lua.LValue {
	uri := L.CheckString(1)

	lh.SetURI(uri)
	pro := treekit.Lazy().Create(L, lh.Name(), typeof)
	if pro.Nil() {
		pro.Set(lh)
	} else {
		_ = pro.Close()
		pro.Set(lh)
	}

	treekit.Start(L, lh, L.PanicErr)
	return pro
}

func (rr *Router) HandleL(L *lua.LState, method string) lua.LValue {
	chain := pipekit.Lua[*webkit.WebContext](L, pipekit.LState(L), pipekit.Protect(true), pipekit.Seek(2))
	h := &LHandle{
		webctx: webkit.NewContext(),
		cfg: config{
			method: method,
		},
		trr:    rr,
		co:     L.Coroutine(),
		handle: chain,
	}
	return h.NewSRV(L)
}
