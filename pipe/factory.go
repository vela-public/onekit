package pipe

import (
	"github.com/vela-public/onekit/lua"
)

type Factory struct {
	Data  any
	Value any
}

func (f *Factory) String() string                         { return "factory" }
func (f *Factory) Type() lua.LValueType                   { return lua.LTObject }
func (f *Factory) AssertFloat64() (float64, bool)         { return float64(0), false }
func (f *Factory) AssertString() (string, bool)           { return "", false }
func (f *Factory) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (f *Factory) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (f *Factory) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "raw":
		return lua.ReflectTo(f.Data)
	case "value":
		return lua.ReflectTo(f.Value)
	}
	return lua.LNil
}

func (f *Factory) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch key {
	case "value":
		f.Value = val
	}
}
