package ssckit

import (
	"context"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/treekit"
	"github.com/vela-public/onekit/tunnel"
	"github.com/vela-public/onekit/zapkit"
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Application struct {
	name    string
	config  *Config
	private struct {
		Context    context.Context
		Cancel     context.CancelFunc
		Executable string
		WorkingDir string
		ServTree   *treekit.MsTree
		TaskTree   *treekit.TaskTree
		Transport  *tunnel.Transport
		Luakit     *luakit.Kit
		Logger     *zapkit.Logger
		Status     layer.StatusBit
	}

	cache struct {
		mutex sync.Mutex
		pool  []layer.Closer
	}

	storage struct {
		ssc *Database
		shm *Database
	}
}

func (app *Application) ServiceTree() layer.ServiceTreeType {
	return app.private.ServTree
}

func (app *Application) init() {
	//init executable path
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	dir, err := filepath.Abs(exe)
	if err != nil {
		panic(err)
	}
	app.private.Executable = exe
	app.private.WorkingDir = dir
}

func (app *Application) NewServTree() {
	ms := treekit.NewMicoServiceOption()
	ms.Protect(true)
	ms.Error(func(e error) {
		app.Error(e)
	})
	ms.Report(func(v *treekit.MsTree) {
		dat := v.View()
		_ = app.Transport().Push("/api/v1/broker/task/status", dat)
	})

	tree := treekit.NewMicoSrvTree(app.Context(), app.private.Luakit, ms)
	tree.Define(app.Transport().R())
	app.private.ServTree = tree
}

func (app *Application) NewTaskTree() {

	option := treekit.NewTaskTreeOption()
	option.Protect(true)
	option.Error(func(err error) {
		app.Error(err)
	})

	option.Report(func(task *treekit.Task) {
		dat := task.Reply()
		err := app.Transport().Push("/api/v1/broker/task/report", dat)
		if err != nil {
			app.Error(err)
		}
	})

	tree := treekit.NewTaskTree(app.Context(), app.private.Luakit, option)
	tree.Define(app.Transport().R())
	app.private.TaskTree = tree
}

func (app *Application) Kill(signal os.Signal) {

}

func (app *Application) Register(cc layer.Closer) {
	app.cache.mutex.Lock()
	defer app.cache.mutex.Unlock()
	app.cache.pool = append(app.cache.pool, cc)
}

func (app *Application) DB() *bbolt.DB {
	if app.storage.ssc == nil {
		return nil
	}

	return app.storage.ssc.dbless
}

func (app *Application) SHM() *bbolt.DB {
	if app.storage.shm == nil {
		return nil
	}
	return app.storage.shm.dbless
}

func (app *Application) Dir() string {
	return app.private.WorkingDir
}

func (app *Application) Exe() string {
	return app.private.Executable
}

func (app *Application) Spawn(i int, f func()) error {
	return nil
}

func (app *Application) Notifier() *Notifier {
	return &Notifier{this: app}
}

func (app *Application) Context() context.Context {
	return app.private.Context
}

func (app *Application) Error(i ...any) {
	if app.private.Logger == nil {
		return
	}

	app.private.Logger.Error(i...)
}

func (app *Application) Errorf(s string, i ...any) {
	if app.private.Logger == nil {
		return
	}
	app.private.Logger.Errorf(s, i...)
}

func (app *Application) Logger() layer.LoggerType {
	if app.private.Logger == nil {
		return nil
	}
	return app.private.Logger
}

func (app *Application) Transport() layer.Transport {
	return app.private.Transport
}

func (app *Application) Name() string {
	return app.name
}

func (app *Application) Devel(vip, edition, host string) {
	ctx := app.Context()
	app.private.Transport.Devel(ctx, vip, edition, host,
		tunnel.WithLogger(app.Logger()),
		tunnel.WithInterval(15*time.Second),
		tunnel.WithNotifier(app.Notifier()))
}

func (app *Application) Node() layer.NodeType {
	return nil
}

func (app *Application) ID() string {
	if app.private.Transport == nil {
		return ""
	}
	return app.private.Transport.ID()
}

func (app *Application) IP() string {

	if app.private.Transport == nil {
		return ""
	}

	if app.private.Transport.Tunnel == nil {
		return ""
	}

	return app.private.Transport.Tunnel.Inet().String()
}

func (app *Application) Preload(v ...func(preloader lua.Preloader)) {
	for _, set := range v {
		set(app.private.Luakit)
	}

}

func (app *Application) Luakit(v ...func(lua.Preloader)) {
	if app.private.Luakit != nil {
		panic("ssoc preload luakit already ok....")
		return
	}
	ps := append(v, app.private.Transport.Preload)
	kit := luakit.Apply(app.name, ps...)
	app.private.Luakit = kit
}

func (app *Application) Prefix() string {
	return app.config.Node.Prefix
}

func (app *Application) With(v ...func(layer.Environment)) {
	for _, set := range v {
		set(app)
	}
}

func (app *Application) Wait() {
	libkit.Wait()
}

func (app *Application) open() {
	app.storage.ssc = &Database{
		name: "ssc",
		dir:  app.Dir(),
		opt: &bbolt.Options{
			Timeout:        10 * time.Second,
			NoGrowSync:     false,
			NoSync:         false,
			NoFreelistSync: false,
			FreelistType:   bbolt.FreelistMapType,
		},
		OnError: app.Errorf,
	}
	app.storage.ssc.Open()
	app.storage.ssc.Define(app.Transport().R())

	app.storage.shm = &Database{
		name: "shm",
		dir:  app.Dir(),
		opt: &bbolt.Options{
			Timeout:        30 * time.Second,
			NoGrowSync:     true,
			NoSync:         true,
			NoFreelistSync: true,
			FreelistType:   bbolt.FreelistMapType,
		},
		OnError: app.Errorf,
	}
	app.storage.shm.Open()
	app.storage.ssc.Define(app.Transport().R())
}

func Apply(parent context.Context, name string, setting ...func(*Application)) *Application {
	cfg := DefaultConfig()
	app := &Application{
		name:   name,
		config: cfg,
	}

	ctx, stop := context.WithCancel(parent)
	app.private.Context = ctx
	app.private.Cancel = stop

	//初始化transport
	app.private.Transport = tunnel.NewTransport(ctx)
	for _, set := range setting {
		set(app)
	}

	app.init()
	app.open()
	app.startup()

	layer.Apply(app)
	return app
}
