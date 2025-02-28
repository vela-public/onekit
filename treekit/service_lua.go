package treekit

import (
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"go.uber.org/zap/zapcore"
	"strings"
)

func (ms *MicroService) setLevel(L *lua.LState) int {
	i := L.Get(1)
	switch i.Type() {
	case lua.LTString:
		level, _ := zapcore.ParseLevel(strings.ToUpper(i.String()))
		ms.setting.Level = level
	case lua.LTInt:
		ms.setting.Level = zapcore.Level(i.(lua.LInt))
	default:
		ms.setting.Level = zapcore.ErrorLevel
	}
	return 0
}

func (ms *MicroService) keepaliveL(L *lua.LState) int {
	ms.setting.Keepalive = lua.IsTrue(L.Get(1))
	return 0
}

func (ms *MicroService) TypeForL(L *lua.LState) int {
	return ms.NewTypeForL(L)
}

func (ms *MicroService) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "level":
		return lua.NewFunction(ms.setLevel)
	case "error":
		return lua.NewFunction(ms.NewServiceErrorL)
	case "debug":
		return lua.NewFunction(ms.NewServiceDebugL)
	case "keepalive":
		return lua.NewFunction(ms.keepaliveL)
	case "T":
		return lua.NewFunction(ms.TypeForL)
	case "GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE", "PATCH":
		return lua.NewFunction(func(co *lua.LState) int {
			r := layer.LazyEnv().Transport().R()
			h := r.HandleL(co, key)
			L.Push(h)
			return 1
		})
	}
	return lua.LNil
}

func (ms *MicroService) startL(L *lua.LState) int {
	p := lua.Check[*Process](L, L.Get(1))
	Start(L, p.data, func(e error) {
		L.RaiseError("%v", e)
	})
	return 0
}

func (ms *MicroService) privateL(L *lua.LState) int {
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		srv := luakit.Check[*Process](L, L.Get(i))
		srv.Private(L)
	}
	return 0
}

func (ms *MicroService) disableL(L *lua.LState) int {
	ms.put(Disable)
	L.RaiseError("disable")
	return 0
}

func (ms *MicroService) importL(L *lua.LState) int {
	dst := strings.Split(L.CheckString(1), ".")
	if len(dst) != 2 {
		L.RaiseError("import name is empty")
		return 0
	}

	key := dst[0]
	name := dst[1]

	tas, ok := ms.root.find(key)
	if !ok {
		L.RaiseError("not found %s", key)
		return 0
	}

	if tas.Key() == ms.Key() {
		L.RaiseError("loop call %s", ms.Key())
		return 0
	}
	ms.link(key)
	//提前唤醒
	if tas.has(Register) {
		_ = tas.wakeup()
	}

	pro, ok := tas.have(name)
	if !ok {
		L.RaiseError("not found %s", key)
		return 0
	}

	L.Push(pro)
	return 1
}

func (ms *MicroService) Preload(kit *luakit.Kit) { //luakit.error() luakit.trace() luakit.T()
	kit.Set("private", lua.NewFunction(ms.privateL))
	kit.Set("disable", lua.NewFunction(ms.disableL))
	kit.Set("event", lua.NewFunction(ms.NewServiceEventL))
	kit.Set("trace", lua.NewFunction(ms.NewServiceTraceL))
	kit.Set("error", lua.NewFunction(ms.NewServiceErrorL))
	kit.Set("debug", lua.NewFunction(ms.NewServiceDebugL))
	kit.Set("T", lua.NewFunction(ms.NewTypeForL))
	kit.SetGlobal("this", lua.NewGeneric[*MicroService](ms))
	kit.SetGlobal("import", lua.NewExport("lua.taskit.export", lua.WithFunc(ms.importL)))
}
