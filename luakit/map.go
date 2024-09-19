package luakit

import (
	"encoding/json"
	"github.com/vela-public/onekit/lua"
)

type Map[K comparable, V any] map[K]V

func (m Map[K, V]) Type() lua.LValueType                   { return lua.LTMap }
func (m Map[K, V]) AssertFloat64() (float64, bool)         { return float64(len(m)), true }
func (m Map[K, V]) AssertString() (string, bool)           { return "", false }
func (m Map[K, V]) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (m Map[K, V]) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (m Map[K, V]) Set(key K, val V) {
	m[key] = val
}

func (m Map[K, V]) String() string {
	text, err := json.Marshal(m)
	if err != nil {
		//todo
		return "{}"
	}
	return lua.B2S(text)
}

func (m Map[K, V]) Meta(L *lua.LState, key lua.LValue) lua.LValue {

	k, ok := key.(K)
	if !ok {
		return lua.LNil
	}

	v, ok := m[k]
	if ok {
		return lua.ToLValue(v)
	}

	return lua.LNil
}

func (m Map[K, V]) Index(L *lua.LState, key string) lua.LValue {
	return m.Meta(L, lua.S2L(key))
}

func (m Map[K, V]) NewIndex(L *lua.LState, key string, value lua.LValue) {
	m.NewMeta(L, lua.S2L(key), value)
}

func NewMap[K comparable, V any](size int) Map[K, V] {
	if size <= 0 {
		return make(Map[K, V])
	} else {
		return make(Map[K, V], size)
	}
}

func (m Map[K, V]) NewMeta(L *lua.LState, key lua.LValue, val lua.LValue) {
	k, ok := key.(K)
	if !ok {
		return
	}
	v, ok := val.(V)
	if !ok {
		return
	}
	m[k] = v
}

func (m Map[K, V]) Decode(data []byte) error {
	return json.Unmarshal(data, &m)
}

func NewMapL(L *lua.LState) int {
	tab := L.CheckTable(1)
	m := Map[lua.LValue, lua.LValue]{}
	tab.ForEach(func(key lua.LValue, value lua.LValue) {
		m.Set(key, value)
	})
	L.Push(lua.NewGeneric[Map[lua.LValue, lua.LValue]](m))
	return 1
}
