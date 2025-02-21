package treekit

import (
	"github.com/vela-public/onekit/event"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"go.uber.org/zap/zapcore"
)

func (ms *MicroService) NewServiceEventL(L *lua.LState) int {
	ev := ms.LazyEvent().Create("logger")
	L.Push(ev)
	return 1
}

func (ms *MicroService) NewServiceTraceL(L *lua.LState) int {
	line := L.Where(1)
	text := "[" + line[:len(line)-1] + "]" + luakit.Format(L, 0)

	ms.LazyEvent().Trace(text).Report()
	return 0
}

func (ms *MicroService) NewServiceErrorL(L *lua.LState) int {
	line := L.Where(1)
	text := "[" + line[:len(line)-1] + "]" + luakit.Format(L, 0)

	ms.LazyEvent().Error(text).Report()
	return 0
}

func (ms *MicroService) LazyEvent() *event.LazyEvent {
	le := event.Lazy(func(ev *event.Event) {
		ev.FromCode = ms.Key()
	})
	return le
}

func (ms *MicroService) NewServiceDebugL(L *lua.LState) int {
	line := L.Where(1)
	text := "[" + line[:len(line)-1] + "]" + luakit.Format(L, 0)

	ev := ms.LazyEvent().Debug(text).Debug()
	if ms.Enable(zapcore.DebugLevel) {
		ev.Report()
	}

	return 0
}

func (ms *MicroService) NewTypeForL(L *lua.LState) int {
	ev := ms.LazyEvent().Debug(luakit.TypeFor(L))
	if ms.Enable(zapcore.DebugLevel) {
		ev.Report()
	}
	return 0
}
