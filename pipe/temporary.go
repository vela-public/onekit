package pipe

import (
	"github.com/vela-public/onekit/lua"
)

type Temporary struct {
	Data  any
	Value any
}

func (t *Temporary) String() string                         { return "temporary" }
func (t *Temporary) Type() lua.LValueType                   { return lua.LTObject }
func (t *Temporary) AssertFloat64() (float64, bool)         { return float64(0), false }
func (t *Temporary) AssertString() (string, bool)           { return "", false }
func (t *Temporary) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (t *Temporary) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (t *Temporary) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "raw":
		return lua.ReflectTo(t.Data)
	case "value":
		return lua.ReflectTo(t.Value)
	}
	return lua.LNil
}

func (t *Temporary) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "value":
		t.Value = val
	}
}
