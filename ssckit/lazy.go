package ssckit

import (
	"context"
	"fmt"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/zapkit"
)

type LazySSOC struct {
	context context.Context
	name    string
	app     *Application
}

func Lazy(name string) *LazySSOC {
	return &LazySSOC{
		name:    name,
		context: context.Background(),
	}
}

func (l *LazySSOC) Context() context.Context {
	return l.context
}

func (l *LazySSOC) stdio(p lua.Preloader) {
	p.SetGlobal("print", lua.NewFunction(func(co *lua.LState) int {
		text := luakit.Format(co, 0)
		fmt.Println(text)
		//l.app.Logger().Debug(text)
		return 0
	}))
}

func (l *LazySSOC) Luakit(v ...func(lua.Preloader)) {
	//加载模块动作
	l.app.Luakit(append(v, l.stdio)...)
}

func (l *LazySSOC) With(v ...func(layer.Environment)) {
	for _, fn := range v {
		fn(l.app)
	}
}

func (l *LazySSOC) Debug(protect bool) {
	// 注入组件
	l.app = Apply(l.Context(), l.name, Protect(protect),
		Logger(zapkit.Debug(zapkit.Console(), zapkit.Caller(2, true))))

}
func (l *LazySSOC) Begin(peer string, other ...string) {

	version := func() string {
		if len(other) < 1 {
			return "4.0.0"
		}
		return other[0]
	}

	hostname := func() string {
		if len(other) < 2 {
			return "vela-ssoc.eastmoney.com"
		}
		return other[1]
	}

	//生成服务管理
	l.app.NewServTree()

	//生成任务管理
	l.app.NewTaskTree()

	//链接中心端
	l.app.Devel(peer, version(), hostname())

	//等待结束信号
	l.app.Wait()

}
