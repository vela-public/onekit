package web

import (
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/treekit"
	"github.com/vela-public/onekit/webkit"
)

type HttpSrv struct {
	cnf *Config
	srv *fasthttp.Server
}

func (hs *HttpSrv) HttpFunc(r *fasthttp.RequestCtx) {
	ctx := webkit.NewWebContext(r)
	hs.cnf.Before.Invoke(ctx)
	hs.cnf.Router.HandlerFunc(r)
	hs.cnf.After.Invoke(ctx)
}

func (hs *HttpSrv) Name() string {
	return hs.cnf.Name
}

func (hs *HttpSrv) Close() error {
	if hs.srv != nil {
		return hs.srv.Shutdown()
	}
	return nil
}

func (hs *HttpSrv) Startup(env *treekit.Env) (err error) {
	hs.srv = &fasthttp.Server{
		Handler: hs.HttpFunc,
	}
	go func() {
		err = hs.srv.ListenAndServe(hs.cnf.Bind)
	}()
	return
}

func NewSrv(cnf *Config) *HttpSrv {
	w := &HttpSrv{
		cnf: cnf,
	}
	return w
}
