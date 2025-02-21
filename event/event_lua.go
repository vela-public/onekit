package event

import (
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
	"time"
)

func (e *Event) String() string                         { return cast.B2S(e.Byte()) }
func (e *Event) Type() lua.LValueType                   { return lua.LTObject }
func (e *Event) AssertFloat64() (float64, bool)         { return float64(len(e.Metadata)), true }
func (e *Event) AssertString() (string, bool)           { return "", false }
func (e *Event) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (e *Event) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (e *Event) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "time":
		e.Time = time.Time(lua.Check[lua.Time](L, val))
	case "subject":
		e.Subject = val.String()
	case "message":
		e.Message = val.String()
	case "alert":
		e.Alert = lua.IsTrue(val)
	case "typeof":
		e.TypeOf = val.String()
	case "level":
		e.Level = val.String()
	case "metadata":
		tab := lua.Check[*lua.LTable](L, val)
		tab.ForEach(func(key lua.LValue, val lua.LValue) {
			e.Metadata[key.String()] = val
		})
	}
}

func (e *Event) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "error":
		e.Error()
	case "debug":
		e.Debug()
	case "report":
		e.Report()
	}

	return e
}
