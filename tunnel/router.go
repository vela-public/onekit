package tunnel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/jsonkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/problem"
	tun "github.com/vela-ssoc/vela-tunnel"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Router struct {
	mutex  sync.Mutex
	cache  map[string]fasthttp.RequestHandler
	route  *router.Router
	inner  *fasthttputil.InmemoryListener
	client *http.Client
}

func (rr *Router) H2S() tun.Server {
	return rr.h2s()
}

func (rr *Router) Serve() *fasthttp.Server {
	return &fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		rr.route.Handler(ctx)
	}}
}

func (rr *Router) Cli() http.Client {
	return *rr.client
}

func (rr *Router) url(req string) string {
	return fmt.Sprintf("http://ssoc/%s", req)
}

func (rr *Router) Exec(method, req string, v interface{}) (*http.Response, error) { //get
	switch method {
	case "GET":
		return rr.client.Get(rr.url(req))
	case "POST":
		return rr.Call(req, v)
	default:
		return nil, fmt.Errorf("method %s not support", method)
	}
}

func (rr *Router) Call(req string, v interface{}) (*http.Response, error) { //post
	switch data := v.(type) {
	case nil:
		return nil, nil
	case io.Reader:
		return rr.client.Post(rr.url(req), "application/json", data)
	case string:
		reader := strings.NewReader(data)
		return rr.client.Post(rr.url(req), "application/json", reader)
	case []byte:
		reader := bytes.NewReader(data)
		return rr.client.Post(rr.url(req), "application/json", reader)
	case fmt.Stringer:
		reader := strings.NewReader(data.String())
		return rr.client.Post(rr.url(req), "application/json", reader)
	case uint8, uint16, uint32, uint, uint64, int8, int16, int32, int, int64, float64, float32:
		reader := strings.NewReader(cast.ToString(data))
		return rr.client.Post(rr.url(req), "application/json", reader)

	default:
		chunk, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		reader := bytes.NewReader(chunk)
		return rr.client.Post(rr.url(req), "application/json", reader)
	}

}

func (rr *Router) Listen() (err error) {
	rr.inner = fasthttputil.NewInmemoryListener()

	go func() {
		err = fasthttp.Serve(rr.inner, func(ctx *fasthttp.RequestCtx) {
			rr.route.Handler(ctx)
		})
	}()

	rr.client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return rr.inner.Dial()
			},
		},
	}
	return
}

func (rr *Router) Bad(ctx *fasthttp.RequestCtx, code int, opt ...func(*problem.Problem)) {
	p := problem.Problem{
		Status:   code,
		Instance: string(ctx.Request.RequestURI()),
	}

	if len(opt) > 0 {
		for _, fn := range opt {
			fn(&p)
		}
	}

	body, _ := json.Marshal(p)
	ctx.Response.SetStatusCode(code)
	ctx.Write(body)
}

func (rr *Router) handler() func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		rr.route.Handler(ctx)
	}
}

func (rr *Router) h2s() tun.Server {
	return &fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		rr.route.Handler(ctx)
	}}
}

func (rr *Router) reload() {
	r := router.New()
	for key, handle := range rr.cache {
		switch {
		case strings.HasPrefix(key, fasthttp.MethodGet):
			r.GET(key[4:], handle)
			continue
		case strings.HasPrefix(key, fasthttp.MethodPost):
			r.POST(key[5:], handle)
			continue
		case strings.HasPrefix(key, fasthttp.MethodDelete):
			r.DELETE(key[7:], handle)
			continue
		case strings.HasPrefix(key, fasthttp.MethodPut):
			r.PUT(key[4:], handle)
			continue
		default:
			//rr.log.Errorf("%s not allow", key)
		}
	}

	rr.route = r
}

func (rr *Router) Upsert(method string, path string, handle fasthttp.RequestHandler) error {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()
	key := fmt.Sprintf("%s_%s", method, path)
	_, ok := rr.cache[key]
	if !ok {
		rr.cache[key] = handle
		rr.route.Handle(method, path, handle)
		return nil
	}

	if !strings.HasPrefix(path, "/api/v1/arr/lua/") {
		return fmt.Errorf("%s %s already ok", method, path)
	}

	rr.cache[key] = handle
	rr.reload()
	return nil
}

func (rr *Router) Then(fn func(*fasthttp.RequestCtx) error) func(*fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		if err := fn(ctx); err != nil {
			rr.Bad(ctx, fasthttp.StatusInternalServerError, problem.Title("内部错误"), problem.Detail(err.Error()))
		}
	}
}

func (rr *Router) Handle(method string, path string, handle fasthttp.RequestHandler) error {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	key := fmt.Sprintf("%s_%s", method, path)

	_, ok := rr.cache[key]
	if ok {
		return fmt.Errorf("%s %s already ok", method, path)
	}
	rr.cache[key] = handle
	rr.route.Handle(method, path, handle)
	return nil
}

func (rr *Router) GET(path string, handle fasthttp.RequestHandler) error {
	return rr.Handle(fasthttp.MethodGet, path, handle)
}

func (rr *Router) POST(path string, handle fasthttp.RequestHandler) error {
	return rr.Handle(fasthttp.MethodPost, path, handle)
}

func (rr *Router) DELETE(path string, handle fasthttp.RequestHandler) error {
	return rr.Handle(fasthttp.MethodDelete, path, handle)
}

func (rr *Router) PUT(path string, handle fasthttp.RequestHandler) error {
	return rr.Handle("put", path, handle)
}

func (rr *Router) Undo(method string, path string) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()
	key := fmt.Sprintf("%s_%s", method, path)

	_, ok := rr.cache[key]
	if !ok {
		return
	}

	delete(rr.cache, key)

	rr.reload()
}

func (rr *Router) callL(L *lua.LState) int {
	return 0
}

func (rr *Router) view(ctx *fasthttp.RequestCtx) error {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()
	enc := jsonkit.NewJson()
	enc.Arr("")
	add := func(method, path string) {
		enc.Tab("")
		enc.KV("method", method)
		enc.KV("full", path)
		enc.KV("path", path[7:])
		enc.End("},")
	}
	for key, _ := range rr.cache {
		switch {
		case strings.HasPrefix(key, fasthttp.MethodGet):
			add(key[:3], key[4:])
			continue
		case strings.HasPrefix(key, fasthttp.MethodPost):
			add(key[:4], key[5:])
			continue
		case strings.HasPrefix(key, fasthttp.MethodDelete):
			add(key[:6], key[7:])
			continue
		case strings.HasPrefix(key, fasthttp.MethodPut):
			add(key[:3], key[4:])
			continue
		default:
			//
		}
	}

	enc.End("]")
	ctx.Write(enc.Bytes())
	return nil

}

func NewRouter() *Router {
	r := &Router{
		cache: make(map[string]fasthttp.RequestHandler, 32),
		route: router.New(),
	}
	r.GET("/api/v1/arr/agent/router/info", r.Then(r.view))
	return r

}
