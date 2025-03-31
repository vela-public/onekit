package pprof

import (
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"net/http"
	"net/http/pprof"
	"time"
)

type Server struct {
	cfg struct {
		Bind string `lua:"bind"`
	}
	err  error
	flag errkit.ErrNo
}

func (srv *Server) String() string                    { return "pprof" }
func (srv *Server) Type() lua.LValueType              { return lua.LTObject }
func (srv *Server) AssertFloat64() (float64, bool)    { return 0, false }
func (srv *Server) AssertString() (string, bool)      { return "", false }
func (srv *Server) Hijack(fsm *lua.CallFrameFSM) bool { return false }

func (srv *Server) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(srv.builder), true
}

func (srv *Server) Logger() layer.LoggerType {
	//todo
	return layer.Logger()
}

func (srv *Server) Errorf(format string, v ...any) {
	l := layer.Logger()
	if l == nil {
		return
	}
	l.Errorf(format, v...)
}

func (srv *Server) start() {
	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprof.Index(w, r)
	}))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprof.Cmdline(w, r)
	}))

	mux.Handle("/debug/pprof/profile", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprof.Profile(w, r)
	}))

	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprof.Symbol(w, r)
	}))

	mux.Handle("/debug/pprof/trace", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprof.Trace(w, r)
	}))

	var err error
	go func() {
		err = http.ListenAndServe(srv.cfg.Bind, mux)
	}()
	time.Sleep(100 * time.Millisecond)
	if err != nil {
		srv.Logger().Errorf("pprof web api start fail %v", err)
		return
	}
	srv.Logger().Errorf("pporf web start http://%s", srv.cfg.Bind)
}

func (srv *Server) builder(L *lua.LState) int {
	if !srv.flag.Have(errkit.Undefined) {
		srv.Errorf("pprof server already start")
		return 0
	}

	tab := L.CheckTable(1)
	err := luakit.TableTo(L, tab, &srv.cfg)
	if err != nil {
		srv.Errorf("pprof create fail %v", err)
		return 0
	}
	srv.start()
	srv.flag = errkit.Succeed
	return 0
}

func Preload(p lua.Preloader) {
	p.Set("pprof", lua.NewExport("lua.pprof.export"))
}
