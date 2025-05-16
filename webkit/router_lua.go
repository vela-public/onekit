package webkit

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipe"
	"net/http/pprof"
)

func (r *Router) String() string                         { return "fasthttp.router" }
func (r *Router) Type() lua.LValueType                   { return lua.LTObject }
func (r *Router) AssertFloat64() (float64, bool)         { return 0, false }
func (r *Router) AssertString() (string, bool)           { return "", false }
func (r *Router) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (r *Router) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (r *Router) NewExecL(L *lua.LState, method string) int {
	uri := L.CheckString(1)
	handle := pipe.Lua(L, pipe.LState(L), pipe.Seek(2))
	r.r.Handle(method, uri, func(ctx *fasthttp.RequestCtx) {
		hc := NewWebContext(ctx)
		handle.Invoke(hc)
	})
	return 0
}

func (r *Router) NotFoundL(L *lua.LState) int {
	handle := pipe.Lua(L, pipe.LState(L), pipe.Seek(1))
	r.r.NotFound = func(ctx *fasthttp.RequestCtx) {
		hc := NewWebContext(ctx)
		handle.Invoke(hc)
	}
	return 0
}

func (r *Router) NewPprofL(L *lua.LState) int {
	base := L.CheckString(1)
	r.GET(base+"/debug/pprof/", func(ctx *fasthttp.RequestCtx) {
		fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Index)(ctx)
	})

	r.GET(base+"/debug/pprof/cmdline", func(ctx *fasthttp.RequestCtx) {
		fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Cmdline)(ctx)
	})
	r.GET(base+"/debug/pprof/profile", func(ctx *fasthttp.RequestCtx) {
		fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Profile)(ctx)
	})
	r.GET(base+"/debug/pprof/symbol", func(ctx *fasthttp.RequestCtx) {
		fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Symbol)(ctx)
	})
	r.GET(base+"/debug/pprof/trace", func(ctx *fasthttp.RequestCtx) {
		fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Trace)(ctx)
	})

	r.GET(base+"/debug/pprof/{name:*}", func(ctx *fasthttp.RequestCtx) {
		uv := ctx.UserValue("name")
		name, ok := uv.(string)
		if !ok {
			ctx.Error("not found", fasthttp.StatusNotFound)
			return
		}
		fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Handler(name).ServeHTTP)(ctx)
	})
	return 0
}

func (r *Router) OnPanicL(L *lua.LState) int {
	handle := pipe.Lua(L, pipe.LState(L), pipe.Seek(1))
	r.r.PanicHandler = func(ctx *fasthttp.RequestCtx, err any) {
		s := lua.NewSlice[lua.LValue](2)
		s.Append(NewWebContext(ctx))
		s.Append(lua.S2L(fmt.Sprintf("%v", err)))
		handle.Invoke(s)
	}
	return 0
}

func (r *Router) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE":
		return lua.NewFunction(func(L *lua.LState) int { return r.NewExecL(L, key) })
	case "ANY":
		return lua.NewFunction(func(L *lua.LState) int { return r.NewExecL(L, "*") })
	case "not_found":
		return lua.NewFunction(r.NotFoundL)
	case "on_panic":
		return lua.NewFunction(r.OnPanicL)

	}
	return lua.LNil
}

func NewHttpRouterL(L *lua.LState) int {
	r := NewRouter()
	L.Push(r)
	return 1
}
