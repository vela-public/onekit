package simple

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/problem"
	"go.etcd.io/bbolt"
	"golang.org/x/net/context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type Simple struct {
	name   string
	luakit *luakit.Kit

	private struct {
		Context    context.Context
		Cancel     context.CancelFunc
		Executable string
		WorkingDir string
	}

	deputy struct {
		Node      layer.NodeType
		Logger    layer.LoggerType
		Router    layer.RouterType
		Transport layer.Tunneler
		Auxiliary layer.Auxiliary
		Preloader layer.Preloader
	}

	cache struct {
		mutex sync.Mutex
		pool  []layer.Closer
	}

	storage struct {
		compacting uint32
		opt        *bbolt.Options
		ssc        *bbolt.DB
		orm        *bbolt.DB
		shm        *bbolt.DB
	}
}

func (sim *Simple) Dir() string {
	return sim.private.WorkingDir
}

func (sim *Simple) Exe() string {
	return sim.private.Executable
}

func (sim *Simple) init() {
	//init executable path
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	dir, err := filepath.Abs(exe)
	if err != nil {
		panic(err)
	}
	sim.private.Executable = exe
	sim.private.WorkingDir = dir
}

func (sim *Simple) Deadline() (deadline time.Time, ok bool) {
	return sim.private.Context.Deadline()
}

func (sim *Simple) Done() <-chan struct{} {
	return sim.private.Context.Done()
}

func (sim *Simple) Err() error {
	return sim.private.Context.Err()
}

func (sim *Simple) Value(key any) any {
	return sim.private.Context.Value(key)
}

func (sim *Simple) Debug(i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Debug(i...)
}

func (sim *Simple) Info(i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Info(i...)
}

func (sim *Simple) Warn(i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Warn(i...)
}

func (sim *Simple) Error(i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Error(i...)
}

func (sim *Simple) Panic(i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}

	sim.deputy.Logger.Panic(i...)
}

func (sim *Simple) Fatal(i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}

	sim.deputy.Logger.Fatal(i...)
}

func (sim *Simple) Trace(i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Trace(i...)
}

func (sim *Simple) Debugf(s string, i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Debugf(s, i...)
}

func (sim *Simple) Infof(s string, i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Infof(s, i...)
}

func (sim *Simple) Warnf(s string, i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Warnf(s, i...)
}

func (sim *Simple) Errorf(s string, i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Errorf(s, i...)
}

func (sim *Simple) Panicf(s string, i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Panicf(s, i...)
}

func (sim *Simple) Fatalf(s string, i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}
	sim.deputy.Logger.Fatalf(s, i...)
}

func (sim *Simple) Tracef(s string, i ...interface{}) {
	if sim.deputy.Logger == nil {
		return
	}

	sim.deputy.Logger.Tracef(s, i...)
}

func (sim *Simple) GET(path string, handle fasthttp.RequestHandler) error {
	if sim.deputy.Router == nil {
		return nil
	}
	return sim.deputy.Router.GET(path, handle)
}

func (sim *Simple) POST(path string, handle fasthttp.RequestHandler) error {
	if sim.deputy.Router == nil {
		return nil
	}
	return sim.deputy.Router.POST(path, handle)
}

func (sim *Simple) DELETE(path string, handle fasthttp.RequestHandler) error {
	if sim.deputy.Router == nil {
		return nil
	}
	return sim.deputy.Router.DELETE(path, handle)
}

func (sim *Simple) PUT(path string, handle fasthttp.RequestHandler) error {
	if sim.deputy.Router == nil {
		return nil
	}
	return sim.deputy.Router.PUT(path, handle)
}

func (sim *Simple) Handle(method, path string, handle fasthttp.RequestHandler) error {
	if sim.deputy.Router == nil {
		return nil
	}
	return sim.deputy.Router.Handle(method, path, handle)
}

func (sim *Simple) Bad(request *fasthttp.RequestCtx, code int, opt ...func(*problem.Problem)) {
	if sim.deputy.Router == nil {
		return
	}
	sim.deputy.Router.Bad(request, code, opt...)
}

func (sim *Simple) Then(f func(*fasthttp.RequestCtx) error) func(*fasthttp.RequestCtx) {
	if sim.deputy.Router == nil {
		return func(sim *fasthttp.RequestCtx) {
			_ = f(sim)
		}
	}
	return sim.deputy.Router.Then(f)
}

func (sim *Simple) Set(s string, value lua.LValue) {
	if sim.deputy.Preloader == nil {
		return
	}
	sim.deputy.Preloader.Set(s, value)
}

func (sim *Simple) Global(s string, value lua.LValue) {
	if sim.deputy.Preloader == nil {
		return
	}
	sim.deputy.Preloader.Global(s, value)
}

func (sim *Simple) Register(cc layer.Closer) {
	sim.cache.mutex.Lock()
	defer sim.cache.mutex.Unlock()
	sim.cache.pool = append(sim.cache.pool, cc)
}

func (sim *Simple) Name() string {
	return sim.name
}

func (sim *Simple) DB() *bbolt.DB {
	return nil
}

func (sim *Simple) Prefix() string {
	//TODO implement me
	panic("implement me")
}

func (sim *Simple) Mode() string {
	return "simple"
}

func (sim *Simple) Spawn(i int, f func()) error {
	//TODO implement me
	panic("implement me")
}

func (sim *Simple) Notify() {
	//TODO implement me
	panic("implement me")
}

func (sim *Simple) Kill(signal os.Signal) {
	//TODO implement me
	panic("implement me")
}

func (sim *Simple) CPU() float64 {
	//TODO implement me
	panic("implement me")
}

func (sim *Simple) AgentCPU() float64 {
	return 0
}

func (sim *Simple) Context() context.Context {
	return sim.private.Context
}

func (sim *Simple) Background() context.Context {
	return sim.private.Context
}

func (sim *Simple) ID() string {
	return ""
}

func (sim *Simple) Arch() string {
	return runtime.GOARCH
}

func (sim *Simple) Inet() string {
	return ""
}

func (sim *Simple) Mac() string {
	return ""
}

func (sim *Simple) Kernel() string {
	return ""
}

func (sim *Simple) Edition() string {
	return ""
}

func (sim *Simple) LocalAddr() string {
	return ""
}

func (sim *Simple) RemoteAddr() string {
	return ""
}

func (sim *Simple) Ident() layer.Ident {
	return layer.Ident{}
}

func (sim *Simple) Quiet() bool {
	return false
}

func (sim *Simple) Broker() (net.IP, int) {
	if sim.deputy.Transport == nil {
		return nil, 0
	}
	return sim.deputy.Transport.Broker()
}

func (sim *Simple) R() layer.RouterType {
	return sim.deputy.Router
}

func (sim *Simple) Node() string {
	return sim.deputy.Transport.Node()
}

func (sim *Simple) Tags() []string {
	if sim.deputy.Transport == nil {
		return nil
	}
	return sim.deputy.Transport.Tags()
}

func (sim *Simple) Doer(prefix string) (layer.Doer, error) {
	if sim.deputy.Transport == nil {
		return nil, fmt.Errorf("transport not configured got %T", sim.deputy.Transport)
	}
	return sim.deputy.Transport.Doer(prefix)
}

func (sim *Simple) Oneway(path string, reader io.Reader, header http.Header) error {
	if sim.deputy.Transport == nil {
		return fmt.Errorf("transport not configured got %T", sim.deputy.Transport)
	}
	return sim.deputy.Transport.Oneway(path, reader, header)
}

func (sim *Simple) Fetch(path string, reader io.Reader, header http.Header) (*http.Response, error) {
	if sim.deputy.Transport == nil {
		return nil, fmt.Errorf("transport not configured got %T", sim.deputy.Transport)
	}
	return sim.deputy.Transport.Fetch(path, reader, header)
}

func (sim *Simple) JSON(path string, data interface{}, result interface{}) error {
	if sim.deputy.Transport == nil {
		return fmt.Errorf("transport not configured got %T", sim.deputy.Transport)
	}
	return sim.deputy.Transport.JSON(path, data, result)
}

func (sim *Simple) Push(path string, data interface{}) error {
	if sim.deputy.Transport == nil {
		return fmt.Errorf("transport not configured got %T", sim.deputy.Transport)
	}
	return sim.deputy.Transport.Push(path, data)
}

func (sim *Simple) OnConnect(name string, todo func() error) {
	if sim.deputy.Transport == nil {
		return
	}
}

func (sim *Simple) Stream(ctx context.Context, s string, header http.Header) (*websocket.Conn, error) {
	if sim.deputy.Transport == nil {
		return nil, fmt.Errorf("transport not configured got %T", sim.deputy.Transport)
	}

	return sim.deputy.Transport.Stream(ctx, s, header)
}

func (sim *Simple) Attachment(name string) (layer.Attachment, error) {
	if sim.deputy.Transport == nil {
		return nil, fmt.Errorf("transport not configured got %T", sim.deputy.Transport)
	}

	return sim.deputy.Transport.Attachment(name)
}

func (sim *Simple) Cancel() {
	sim.private.Cancel()
}

func New(name string) *Simple {

	ctx, stop := context.WithCancel(context.Background())
	sim := &Simple{name: name}
	sim.private.Context = ctx
	sim.private.Cancel = stop
	return sim
}
