package tunnel

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipe"
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
	co      *lua.LState
	webctx  *webkit.HttpContext
	handle  *pipe.Chain
	router  *Router
}

func (cfg *config) Name() string {
	return fmt.Sprintf("%s %s", cfg.method, cfg.uri)
}

func (cfg *config) SetURI(path string) {
	var uri string
	if !strings.HasPrefix(path, "/api/v1/arr/lua/") {
		uri = filepath.Join("/api/v1/arr/lua/", cfg.co.Name(), path)
	} else {
		uri = filepath.Join("/", cfg.co.Name(), path)
	}
	cfg.uri = filepath.ToSlash(uri)
}

type LHandle struct {
	cfg     *config
	trr     *Router
	private struct {
		Count   int64 //请求次数
		Failed  int64 //失败次数
		Succeed int64 //成功次数
	}
}

func (lh *LHandle) Name() string {
	return lh.cfg.Name()
}

func (lh *LHandle) Metadata() libkit.DataKV[string, any] {
	mt := libkit.NewDataKV[string, any]()
	mt.Set("count", lh.private.Count)
	mt.Set("failed", lh.private.Failed)
	mt.Set("succeed", lh.private.Succeed)
	return *mt
}

func (lh *LHandle) HandleFunc(ctx *fasthttp.RequestCtx) {
	webctx := webkit.NewWebContext(ctx)
	lh.cfg.handle.Invoke(webctx)
}

func (lh *LHandle) Start() error {
	if lh.cfg.handle == nil {
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

func NewHandleL(L *lua.LState, rr *Router, method string) lua.LValue {
	path := L.CheckString(1)
	chain := pipe.Lua(L, pipe.LState(L), pipe.Protect(true), pipe.Seek(2))

	cfg := &config{
		method: method,
		handle: chain,
		router: rr,
	}
	cfg.SetURI(path)

	pro := treekit.LazyCNF[LHandle, config](L, cfg)
	pro.Upsert(func(*config) *LHandle {
		return &LHandle{
			cfg: cfg,
			trr: rr,
		}
	})
	pro.Start()

	return pro.Unwrap()
}
