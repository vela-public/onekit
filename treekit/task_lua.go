package treekit

import (
	"github.com/vela-public/onekit/event"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"go.uber.org/zap/zapcore"
	"strings"
)

func (t *Task) privateL(L *lua.LState) int {
	return 0
}

func (t *Task) disableL(L *lua.LState) int {
	return 0
}

func (t *Task) NewTaskEventL(L *lua.LState) int {
	ev := t.LazyEvent().Create("logger")
	L.Push(ev)
	return 1
}

func (t *Task) NewTaskTraceL(L *lua.LState) int {
	return 0
}

func (t *Task) LazyEvent() *event.LazyEvent {
	le := event.Lazy(func(ev *event.Event) {
		ev.FromCode = t.From()
		ev.Set("id", t.config.ID)
		ev.Set("exec_id", t.config.ExecID)
	})
	return le
}

func (t *Task) NewTaskErrorL(L *lua.LState) int {
	line := L.Where(1)
	text := "[" + line[:len(line)-1] + "]" + luakit.Format(L, 0)

	t.LazyEvent().Error(text).Error().Report()
	return 0

}

func (t *Task) NewTaskDebugL(L *lua.LState) int {
	line := L.Where(1)
	text := "[" + line[:len(line)-1] + "]" + luakit.Format(L, 0)

	ev := t.LazyEvent().Debug(text)
	if t.Enable(zapcore.DebugLevel) {
		ev.Report()
	}
	return 0
}

func (t *Task) NewTypeForL(L *lua.LState) int {
	ev := t.LazyEvent().Trace(luakit.TypeFor(L))
	if t.Enable(zapcore.DebugLevel) {
		ev.Report()
	}

	return 0
}

func (t *Task) importL(L *lua.LState) int {
	srv := layer.LazyEnv().ServiceTree()
	return srv.Lookup(L)
}

func (t *Task) Enable(v zapcore.Level) bool {
	return v >= t.setting.Level
}

func (t *Task) setLevel(L *lua.LState) int {
	i := L.Get(1)
	switch i.Type() {
	case lua.LTString:
		level, _ := zapcore.ParseLevel(strings.ToUpper(i.String()))
		t.setting.Level = level
	case lua.LTInt:
		t.setting.Level = zapcore.Level(i.(lua.LInt))
	default:
		t.setting.Level = zapcore.ErrorLevel
	}
	return 0
}

func (t *Task) keepaliveL(L *lua.LState) int {
	return 0
}

func (t *Task) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "level":
		return lua.NewFunction(t.setLevel)
	case "trace":
		return lua.NewFunction(t.NewTaskTraceL)
	case "keepalive":
		return lua.NewFunction(t.keepaliveL)
	case "event":
		return lua.NewFunction(t.NewTaskEventL)
	case "T":
		return lua.NewFunction(t.NewTypeForL)
	}
	return lua.LNil
}

func (t *Task) Preload(kit *luakit.Kit) { //luakit.error() luakit.trace() luakit.T()
	kit.Set("private", lua.NewFunction(t.privateL))
	kit.Set("disable", lua.NewFunction(t.disableL))
	kit.Set("event", lua.NewFunction(t.NewTaskEventL))
	kit.Set("error", lua.NewFunction(t.NewTaskErrorL))
	kit.Set("debug", lua.NewFunction(t.NewTaskDebugL))
	kit.SetGlobal("this", lua.NewGeneric[*Task](t))
	kit.SetGlobal("import", lua.NewExport("lua.taskit.export", lua.WithFunc(t.importL)))
}
