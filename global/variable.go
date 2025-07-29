package global

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"sync"
)

type Variable struct {
	mutex sync.RWMutex
	cache map[string]any
}

func (va *Variable) String() string                         { return fmt.Sprintf("var[%d]", len(va.cache)) }
func (va *Variable) Type() lua.LValueType                   { return lua.LTObject }
func (va *Variable) AssertFloat64() (float64, bool)         { return float64(len(va.cache)), true }
func (va *Variable) AssertString() (string, bool)           { return "", false }
func (va *Variable) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (va *Variable) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (va *Variable) Index(L *lua.LState, key string) lua.LValue {
	va.mutex.RLock()
	defer va.mutex.RUnlock()

	dat, ok := va.cache[key]
	if !ok {
		return lua.LNil
	}

	lv, ok := lua.TypeFor(dat)
	if !ok {
		return lua.LNil
	}
	return lv
}

func (va *Variable) NewIndex(L *lua.LState, key string, val lua.LValue) {
	va.mutex.Lock()
	defer va.mutex.Unlock()
	va.cache[key] = val
}
func (va *Variable) MetaTable(L *lua.LState, key string) lua.LValue {
	return va.Index(L, key)
}

func NewVariable() *Variable {
	return &Variable{
		cache: make(map[string]any),
	}
}
