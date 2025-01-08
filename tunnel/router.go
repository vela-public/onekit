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

type TRouter struct {
	mutex  sync.Mutex
	cache  map[string]fasthttp.RequestHandler
	route  *router.Router
	inner  *fasthttputil.InmemoryListener
	client *http.Client
}

func (trr *TRouter) H2S() tun.Server {
	return trr.h2s()
}

func (trr *TRouter) Serve() *fasthttp.Server {
	return &fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		trr.route.Handler(ctx)
	}}
}

func (trr *TRouter) Cli() http.Client {
	return *trr.client
}

func (trr *TRouter) url(req string) string {
	return fmt.Sprintf("http://ssoc/%s", req)
}

func (trr *TRouter) Exec(method, req string, v interface{}) (*http.Response, error) { //get
	switch method {
	case "GET":
		return trr.client.Get(trr.url(req))
	case "POST":
		return trr.Call(req, v)
	default:
		return nil, fmt.Errorf("method %s not support", method)
	}
}

func (trr *TRouter) Call(req string, v interface{}) (*http.Response, error) { //post
	switch data := v.(type) {
	case nil:
		return nil, nil
	case io.Reader:
		return trr.client.Post(trr.url(req), "application/json", data)
	case string:
		reader := strings.NewReader(data)
		return trr.client.Post(trr.url(req), "application/json", reader)
	case []byte:
		reader := bytes.NewReader(data)
		return trr.client.Post(trr.url(req), "application/json", reader)
	case fmt.Stringer:
		reader := strings.NewReader(data.String())
		return trr.client.Post(trr.url(req), "application/json", reader)
	case uint8, uint16, uint32, uint, uint64, int8, int16, int32, int, int64, float64, float32:
		reader := strings.NewReader(cast.ToString(data))
		return trr.client.Post(trr.url(req), "application/json", reader)

	default:
		chunk, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		reader := bytes.NewReader(chunk)
		return trr.client.Post(trr.url(req), "application/json", reader)
	}

}

func (trr *TRouter) Listen() (err error) {
	trr.inner = fasthttputil.NewInmemoryListener()

	go func() {
		err = fasthttp.Serve(trr.inner, func(ctx *fasthttp.RequestCtx) {
			trr.route.Handler(ctx)
		})
	}()

	trr.client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return trr.inner.Dial()
			},
		},
	}
	return
}

func (trr *TRouter) Bad(ctx *fasthttp.RequestCtx, code int, opt ...func(*problem.Problem)) {
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

func (trr *TRouter) handler() func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		trr.route.Handler(ctx)
	}
}

func (trr *TRouter) h2s() tun.Server {
	return &fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		trr.route.Handler(ctx)
	}}
}

func (trr *TRouter) reload() {
	r := router.New()
	for key, handle := range trr.cache {
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
			//trr.log.Errorf("%s not allow", key)
		}
	}

	trr.route = r
}

func (trr *TRouter) Upsert(method string, path string, handle fasthttp.RequestHandler) error {
	trr.mutex.Lock()
	defer trr.mutex.Unlock()
	key := fmt.Sprintf("%s_%s", method, path)
	_, ok := trr.cache[key]
	if !ok {
		trr.cache[key] = handle
		trr.route.Handle(method, path, handle)
		return nil
	}

	if !strings.HasPrefix(path, "/api/v1/arr/lua/") {
		return fmt.Errorf("%s %s already ok", method, path)
	}

	trr.cache[key] = handle
	trr.reload()
	return nil
}

func (trr *TRouter) Then(fn func(*fasthttp.RequestCtx) error) func(*fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		if err := fn(ctx); err != nil {
			trr.Bad(ctx, fasthttp.StatusInternalServerError, problem.Title("内部错误"), problem.Detail(err.Error()))
		}
	}
}

func (trr *TRouter) Handle(method string, path string, handle fasthttp.RequestHandler) error {
	trr.mutex.Lock()
	defer trr.mutex.Unlock()

	key := fmt.Sprintf("%s_%s", method, path)

	_, ok := trr.cache[key]
	if ok {
		return fmt.Errorf("%s %s already ok", method, path)
	}
	trr.cache[key] = handle
	trr.route.Handle(method, path, handle)
	return nil
}

func (trr *TRouter) GET(path string, handle fasthttp.RequestHandler) error {
	return trr.Handle(fasthttp.MethodGet, path, handle)
}

func (trr *TRouter) POST(path string, handle fasthttp.RequestHandler) error {
	return trr.Handle(fasthttp.MethodPost, path, handle)
}

func (trr *TRouter) DELETE(path string, handle fasthttp.RequestHandler) error {
	return trr.Handle(fasthttp.MethodDelete, path, handle)
}

func (trr *TRouter) PUT(path string, handle fasthttp.RequestHandler) error {
	return trr.Handle("put", path, handle)
}

func (trr *TRouter) Undo(method string, path string) {
	trr.mutex.Lock()
	defer trr.mutex.Unlock()
	key := fmt.Sprintf("%s_%s", method, path)

	_, ok := trr.cache[key]
	if !ok {
		return
	}

	delete(trr.cache, key)

	trr.reload()
}

func (trr *TRouter) callL(L *lua.LState) int {
	return 0
}

func (trr *TRouter) view(ctx *fasthttp.RequestCtx) error {
	trr.mutex.Lock()
	defer trr.mutex.Unlock()
	enc := jsonkit.NewJson()
	enc.Arr("")
	add := func(method, path string) {
		enc.Tab("")
		enc.KV("method", method)
		enc.KV("full", path)
		enc.KV("path", path[7:])
		enc.End("},")
	}
	for key, _ := range trr.cache {
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
