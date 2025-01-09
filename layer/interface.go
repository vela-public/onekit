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

type Transport interface {
	Broker() (net.IP, int)
	R() RouterType
	Node() string
	Tags() []string
	Doer(prefix string) (Doer, error)
	Oneway(path string, reader io.Reader, header http.Header) error
	Fetch(path string, reader io.Reader, header http.Header) (*http.Response, error)
	JSON(path string, data interface{}, result interface{}) error
	Push(path string, data interface{}) error
	Stream(context.Context, string, http.Header) (*websocket.Conn, error)
	Attachment(name string) (Attachment, error)
}

type Environment interface {
	Register(Closer)          //注册关闭器
	Name() string             //当前环境的名称
	ID() string               //当前环境的ID
	IP() string               //当前环境的IP
	DB() *bbolt.DB            //当前环境的缓存库
	Dir() string              //当前环境目录
	Exe() string              //运行executable
	Spawn(int, func()) error  //异步执行 (delay int , task func())
	Context() context.Context //全局context
	Logger() LoggerType       //日志接口
	Node() NodeType           //节点信息
	Transport() Transport     //网络接口
}
