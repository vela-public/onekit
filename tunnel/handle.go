package tunnel

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipekit"
	"github.com/vela-public/onekit/taskit"
	"github.com/vela-public/onekit/webkit"
	"net/http"
	"path/filepath"
	"reflect"
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
	handle  *pipekit.Chain[*fasthttp.RequestCtx]
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

func (lh *LHandle) HandleFunc() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		lh.handle.Invoke(ctx)
	}
}

func (lh *LHandle) Start() error {
	if lh.handle == nil {
		return fmt.Errorf("router start fail not found handle")
	}
	return lh.trr.Handle(lh.cfg.method, lh.cfg.uri, lh.HandleFunc())
}

func (lh *LHandle) Close() error {
	if lh.cfg.recycle {
		return nil
	}

	lh.cfg.recycle = true
	lh.trr.Undo(lh.cfg.method, lh.cfg.uri)
	return nil
}

func (lh *LHandle) ToCall(L *lua.LState) int {

	uri := L.IsString(1)
	handle := L.Get(2)
	if len(uri) < 2 {
		L.RaiseError("invalid router uri got empty")
		return 0
	}

	tas := taskit.CheckTaskEx(L, func(err error) {
		L.RaiseError("%v", err)
	})

	if tas == nil {
		L.RaiseError("not allow add router by not task")
		return 0
	}

	path := filepath.Join("/api/v1/arr/lua/", tas.Key(), uri)

	lh.cfg.uri = filepath.ToSlash(path)

	srv := taskit.Create(L, lh.Name(), typeof)
	if srv.Nil() {
		srv.Set(lh)
	} else {
		_ = srv.Close()
		srv.Set(lh)
	}

	switch handle.Type() {
	case lua.LTNil:
		lh.handle.NewHandler(func(ctx *fasthttp.RequestCtx) {
			_, _ = ctx.WriteString("not found handle")
		})
	case lua.LTString:
		lh.handle.NewHandler(func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString(handle.String())
		})

	case lua.LTFunction:
		fn, ok := handle.AssertFunction()
		if !ok {
			L.RaiseError("invalid function value")
			return 0
		}

		lh.handle.NewHandler(func(ctx *fasthttp.RequestCtx) {
			co := L.Coroutine()
			defer L.Keepalive(co)
			pn := lua.P{
				Fn:   fn,
				NRet: 0,
			}
			co.SetValue(webkit.WEB_CONTEXT_KEY, ctx)

			err := co.CallByParam(pn, lh.webctx)
			if err != nil {
				ctx.Error(err.Error(), http.StatusInternalServerError)
				return
			}
		})

	default:
		lh.handle.NewHandler(func(ctx *fasthttp.RequestCtx) {
			_, _ = ctx.WriteString(handle.String())
		})
	}

	err := lh.Start()
	if err != nil {
		L.RaiseError("inject router error %v", err)
	}

	return 0
}

func (lh *LHandle) LFunc(L *lua.LState) *lua.LFunction {
	return L.NewFunction(lh.ToCall)
}

func (rr *Router) newHandleL(L *lua.LState, method string) lua.LValue {
	lh := &LHandle{
		webctx: webkit.NewContext(),
		cfg: config{
			method: method,
		},
		trr:    rr,
		co:     L.Coroutine(),
		handle: pipekit.NewChain[*fasthttp.RequestCtx](),
	}

	return lh.LFunc(L)
}
