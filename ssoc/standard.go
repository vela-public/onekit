package ssoc

import (
	"context"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/taskit"
	"github.com/vela-public/onekit/tunnel"
	tun "github.com/vela-ssoc/vela-tunnel"
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Standard struct {
	option  *Options
	private struct {
		Context    context.Context
		Cancel     context.CancelFunc
		Executable string
		WorkingDir string
		Tree       *taskit.Tree
		Transport  *tunnel.Transport
		Luakit     *luakit.Kit
		Logger     layer.LoggerType
		Status     layer.StatusBit
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

func (std *Standard) NewTree() {
	tree := taskit.NewTree(std.Context(), std.private.Luakit,
		//report
		taskit.Report(func(v *taskit.Tree) {
			dat := v.View()
			_ = std.Transport().Push("/api/v1/broker/task/status", dat)
		}),

		//panic protect
		taskit.Protect(std.option.protect),
		//error handle
		taskit.OnError(func(e error) {
			std.Logger().Error(e)
		}))

	tree.Define(std.Transport().R())
	std.private.Tree = tree
}

func (std *Standard) NewPath() {
	//init executable path
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	dir, err := filepath.Abs(exe)
	if err != nil {
		panic(err)
	}
	std.private.Executable = exe
	std.private.WorkingDir = dir
}

func (std *Standard) Kill(signal os.Signal) {

}

func (std *Standard) Register(cc layer.Closer) {
	std.cache.mutex.Lock()
	defer std.cache.mutex.Unlock()
	std.cache.pool = append(std.cache.pool, cc)
}

func (std *Standard) DB() *bbolt.DB {
	return nil
}

func (std *Standard) Dir() string {
	return std.private.WorkingDir
}

func (std *Standard) Exe() string {
	return std.private.Executable
}

func (std *Standard) Spawn(i int, f func()) error {
	return nil
}

func (std *Standard) Notifier() *Notifier {
	return &Notifier{std: std}
}

func (std *Standard) Context() context.Context {
	return std.private.Context
}

func (std *Standard) Logger() layer.LoggerType {
	if std.private.Logger == nil {
		return nil
	}
	return std.private.Logger
}

func (std *Standard) Transport() layer.Transport {
	if std.private.Transport == nil {
		return NoopTransport{}
	}

	return std.private.Transport
}

func (std *Standard) Name() string {
	return std.option.name
}

func (std *Standard) Devel(vip, edition, host string) {
	ctx := std.Context()
	std.private.Transport.Devel(ctx, vip, edition, host,
		tun.WithLogger(std.Logger()),
		tun.WithInterval(15*time.Second),
		tun.WithNotifier(std.Notifier()))
}

func (std *Standard) Layer() {
	layer.Apply(std)
}

func (std *Standard) Luakit(v ...func(lua.Preloader)) {
	ps := append(v, std.private.Transport.Preload)
	kit := luakit.Apply(std.option.name, ps...)
	std.private.Luakit = kit
}

func (std *Standard) OpenDB() {

}

func (std *Standard) Node() layer.NodeType {
	return nil
}

func (std *Standard) ID() string {
	return std.private.Transport.ID()
}

func (std *Standard) IP() string {
	return std.private.Transport.Tunnel.Inet().String()
}

func (std *Standard) Wait() {
	libkit.Wait()
}

func Create(parent context.Context, name string, setting ...func(*Options)) *Standard {
	opt := &Options{
		name: name,
	}

	std := &Standard{
		option: opt,
	}

	for _, set := range setting {
		set(opt)
	}

	ctx, stop := context.WithCancel(parent)
	std.private.Context = ctx
	std.private.Cancel = stop

	//初始化transport
	std.private.Transport = tunnel.NewTransport(ctx)

	std.NewPath()
	return std
}
