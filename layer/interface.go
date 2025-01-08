package layer

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/problem"
	"go.etcd.io/bbolt"
	"io"
	"net"
	"net/http"
	"os"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Attachment interface {
	Filename() string
	Hash() string
	NotModified() bool
	Read([]byte) (int, error)
	Close() error
	ZipFile() bool
	WriteTo(w io.Writer) (int64, error)
	File(path string) (string, error)
	Unzip(path string) error
}

type RouterType interface {
	GET(path string, handle fasthttp.RequestHandler) error
	POST(path string, handle fasthttp.RequestHandler) error
	DELETE(path string, handle fasthttp.RequestHandler) error
	PUT(path string, handle fasthttp.RequestHandler) error
	Handle(method, path string, handle fasthttp.RequestHandler) error
	Bad(ctx *fasthttp.RequestCtx, code int, opt ...func(*problem.Problem))
	Then(func(*fasthttp.RequestCtx) error) func(*fasthttp.RequestCtx)
}

type LoggerType interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Panic(...interface{})
	Fatal(...interface{})
	Trace(...interface{})

	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Panicf(string, ...interface{})
	Fatalf(string, ...interface{})
	Tracef(string, ...interface{})
}

type NodeType interface {
	ID() string
	Arch() string
	Inet() string
	Mac() string
	Kernel() string
	Edition() string
	LocalAddr() string
	RemoteAddr() string
	Ident() Ident
	Quiet() bool
}

type Preloader interface {
	Set(string, lua.LValue)
	Global(string, lua.LValue)
}

type Closer interface {
	Name() string
	Close() error
}

type Auxiliary interface {
	Register(Closer)          //注册关闭器
	Name() string             //当前环境的名称
	DB() *bbolt.DB            //当前环境的缓存库
	Prefix() string           //系统前缀
	Dir() string              //当前环境目录
	Exe() string              //运行executable
	Mode() string             //当前环境模式
	Spawn(int, func()) error  //异步执行 (delay int , task func())
	Notify()                  //监控退出信号
	Kill(os.Signal)           //退出
	Context() context.Context //全局context
}

type Tunneler interface {
	Broker() (net.IP, int)
	R() RouterType
	Node() string
	Tags() []string
	Doer(prefix string) (Doer, error)
	Oneway(path string, reader io.Reader, header http.Header) error
	Fetch(path string, reader io.Reader, header http.Header) (*http.Response, error)
	JSON(path string, data interface{}, result interface{}) error
	Push(path string, data interface{}) error
	OnConnect(name string, todo func() error)
	Stream(context.Context, string, http.Header) (*websocket.Conn, error)
	Attachment(name string) (Attachment, error)
}

type Environment interface {
	context.Context
	LoggerType
	RouterType
	Preloader
	Auxiliary
	NodeType
	Tunneler
}
