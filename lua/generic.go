package lua

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/vela-public/onekit/cast"
)

type Generic[T any] struct {
	Data T
}

func NewGeneric[T any](data T) *Generic[T] {
	return &Generic[T]{
		Data: data,
	}
}

func (gen *Generic[T]) ToLValue() LValue {
	return ToLValue(gen.Data)
}

func (gen *Generic[T]) GobEncode() ([]byte, error) {
	buff := new(bytes.Buffer)
	err := gob.NewEncoder(buff).Encode(&gen.Data)
	return buff.Bytes(), err
}

func (gen *Generic[T]) GobDecode(data []byte) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(&gen.Data)
}

func (gen *Generic[T]) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &gen.Data)
}

func (gen *Generic[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(gen.Data)
}

func (gen *Generic[T]) LValue() (LValue, bool) {
	var v interface{} = gen.Data

	if lv, ok := v.(LValue); ok {
		return lv, true
	}
	return LNil, false
}

func (gen *Generic[T]) String() string {
	if v, ok := gen.LValue(); ok {
		return v.String()
	}

	return cast.ToString(gen.Data)
}

func (gen *Generic[T]) Type() LValueType {
	return LTGeneric
}

func (gen *Generic[T]) AssertFloat64() (float64, bool) {
	if v, ok := gen.LValue(); ok {
		return v.AssertFloat64()
	}

	return 0, false
}

func (gen *Generic[T]) AssertString() (string, bool) {
	if v, ok := gen.LValue(); ok {
		return v.AssertString()
	}
	return "", false
}

func (gen *Generic[T]) AssertFunction() (*LFunction, bool) {

	var a any = gen.Data

	switch vt := a.(type) {
	case *LFunction:
		return vt, true
	case LValue:
		return vt.AssertFunction()
	case interface{ AssertFunction() (*LFunction, bool) }:
		return vt.AssertFunction()
	case interface{ ToLFunction() *LFunction }:
		return vt.ToLFunction(), true

	case func():
		return NewFunction(func(L *LState) int {
			vt()
			return 0
		}), true

	case func() error:
		return NewFunction(func(L *LState) int {
			if err := vt(); err != nil {
				L.Push(S2L(err.Error()))
				return 1
			}
			return 0
		}), true

	}
	if v, ok := gen.LValue(); ok {
		return v.AssertFunction()
	}

	return nil, false
}

func (gen *Generic[T]) Index(L *LState, key string) LValue {
	var value any = gen.Data

	if v, ok := value.(IndexEx); ok {
		return v.Index(L, key)
	}

	return LNil
}

func (gen *Generic[T]) NewIndex(L *LState, key string, val LValue) {
	var value any = gen.Data

	if v, ok := value.(NewIndexEx); ok {
		v.NewIndex(L, key, val)
	}
}

func (gen *Generic[T]) Meta(L *LState, key LValue) LValue {
	var value any = gen.Data
	if v, ok := value.(MetaEx); ok {
		return v.Meta(L, key)
	}
	return LNil
}

func (gen *Generic[T]) NewMeta(L *LState, key LValue, val LValue) {
	var value any = gen.Data
	if v, ok := value.(NewMetaEx); ok {
		v.NewMeta(L, key, val)
		return
	}
}

func (gen *Generic[T]) Hijack(fsm *CallFrameFSM) bool {
	if v, ok := gen.LValue(); ok {
		return v.Hijack(fsm)
	}

	return false
}
