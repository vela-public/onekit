package web

import (
	"github.com/vela-public/onekit/pipe"
	"github.com/vela-public/onekit/webkit"
)

type Config struct {
	Name    string                              `lua:"name"`
	Cluster []string                            `lua:"cluster"`
	Bind    string                              `lua:"bind"`
	Router  *webkit.Router                      `lua:"-"`
	Before  *pipe.LazyChain[*webkit.WebContext] `lua:"-"`
	After   *pipe.LazyChain[*webkit.WebContext] `lua:"-"`
}

func (cnf *Config) Must() {
	cnf.Before = pipe.NewLazyChain[*webkit.WebContext]()
	cnf.After = pipe.NewLazyChain[*webkit.WebContext]()
	cnf.Router = webkit.NewRouter()
}
