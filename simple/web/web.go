package web

import (
	"context"
	"github.com/valyala/fasthttp"
)

type HttpSrv struct {
	cnf *Config
	srv *fasthttp.Server
}

func (hs *HttpSrv) HttpFunc(ctx *fasthttp.RequestCtx) {
	hs.cnf.Router.HandlerFunc(ctx)
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

func (hs *HttpSrv) Start(ctx context.Context) (err error) {
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
