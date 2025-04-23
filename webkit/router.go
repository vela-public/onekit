package webkit

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/pipe"
)

type Router struct {
	r          *router.Router
	Middleware struct {
		OnRequest  *pipe.LazyChain[*WebContext]
		Chain      *pipe.LazyChain[*WebContext]
		Switch     *pipe.LazySwitch[*WebContext]
		OnResponse *pipe.LazyChain[*WebContext]
	}
}

func (r *Router) GET(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodGet, uri, h)
}

func (r *Router) HEAD(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodHead, uri, h)
}

func (r *Router) POST(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodPost, uri, h)
}

func (r *Router) PUT(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodPut, uri, h)
}

func (r *Router) PATCH(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodPatch, uri, h)
}

func (r *Router) DELETE(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodDelete, uri, h)
}

func (r *Router) CONNECT(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodConnect, uri, h)
}

func (r *Router) OPTIONS(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodOptions, uri, h)
}

func (r *Router) TRACE(uri string, h fasthttp.RequestHandler) {
	r.r.Handle(fasthttp.MethodTrace, uri, h)
}

func (r *Router) Handler(method, uri string, h fasthttp.RequestHandler) {
	r.r.Handle(method, uri, h)
}

func (r *Router) HandlerFunc(req *fasthttp.RequestCtx) {
	ctx := NewWebContext(req)
	r.Middleware.Chain.Invoke(ctx)
	r.Middleware.Switch.Invoke(ctx)
	r.Middleware.OnRequest.Invoke(ctx)
	r.r.Handler(req)
	r.Middleware.OnResponse.Invoke(ctx)
}

func NewRouter() *Router {
	r := &Router{r: router.New()}
	r.Middleware.Chain = pipe.NewLazyChain[*WebContext]()
	r.Middleware.OnRequest = pipe.NewLazyChain[*WebContext]()
	r.Middleware.Switch = pipe.NewLazySwitch[*WebContext]()
	r.Middleware.OnResponse = pipe.NewLazyChain[*WebContext]()
	return r
}
