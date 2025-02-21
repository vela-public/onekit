package lua

import (
	"encoding/json"
)

const (
	MapLibName = "map"
)

type Map[K comparable, V any] map[K]V

func (m Map[K, V]) Type() LValueType                   { return LTMap }
func (m Map[K, V]) AssertFloat64() (float64, bool)     { return float64(len(m)), true }
func (m Map[K, V]) AssertString() (string, bool)       { return "", false }
func (m Map[K, V]) AssertFunction() (*LFunction, bool) { return nil, false }
func (m Map[K, V]) Hijack(fsm *CallFrameFSM) bool      { return false }

func (m Map[K, V]) Set(key K, val V) {
	m[key] = val
}

func (m Map[K, V]) String() string {
	text, err := json.Marshal(m)
	if err != nil {
		//todo
		return "{}"
	}
	return B2S(text)
}

func (m Map[K, V]) Meta(L *LState, key LValue) LValue {

	k, ok := key.(K)
	if !ok {
		return LNil
	}

	v, ok := m[k]
	if ok {
		return ReflectTo(v)
	}

	return LNil
}

func (m Map[K, V]) Index(L *LState, key string) LValue {
	return m.Meta(L, S2L(key))
}

func (m Map[K, V]) NewIndex(L *LState, key string, value LValue) {
	m.NewMeta(L, S2L(key), value)
}

func NewMap[K comparable, V any](size int) Map[K, V] {
	if size <= 0 {
		return make(Map[K, V])
	} else {
		return make(Map[K, V], size)
	}
}

func (m Map[K, V]) NewMeta(L *LState, key LValue, val LValue) {
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

func NewMapL(L *LState) int {
	tab := L.CheckTable(1)
	m := Map[LValue, LValue]{}
	tab.ForEach(func(key LValue, value LValue) {
		m.Set(key, value)
	})
	L.Push(m)
	return 1
}

func MapTo[K comparable, V any](m map[K]V) Map[K, V] {
	return m
}

func OpenMapLib(L *LState) int {
	mod := NewExport("lua.map.export", WithFunc(NewMapL))
	L.SetGlobal("map", mod)
	L.Push(mod)
	return 1
}
