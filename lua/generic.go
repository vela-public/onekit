package lua

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/vela-public/onekit/cast"
)

type (
	Call0 func()
	Call1 func(any)
	Call2 func(any, any)
	Call3 func(any, any, any)
)
type (
	CallE0 func() error
	CallE1 func(any) error
	CallE2 func(any, any) error
	CallE3 func(any, any, any) error
)

type ValueType interface {
	ToValue() LValue
}

type FunctionType interface {
	AssertFunction() (*LFunction, bool)
}

type Float64Type interface {
	AssertFloat64() (float64, bool)
}

type StringType interface {
	AssertString() (string, bool)
}

type HijackType interface {
	Hijack(*CallFrameFSM) bool
}

type IndexType interface {
	Index(*LState, string) LValue
}

type Getter interface {
	Getter(string) any
}

type IndexOfType interface {
	IndexOf(*LState, string) LValue
}

type Setter interface {
	Setter(string, any)
}

type NewIndexType interface {
	NewIndex(*LState, string, LValue)
}

type NewIndexOfType interface {
	NewIndexOf(*LState, string, LValue)
}

type MetaType interface {
	Meta(*LState, LValue) LValue
}

type NewMetaType interface {
	NewMeta(*LState, LValue, LValue)
}

type MetaOfType interface {
	MetaOf(*LState, LValue) LValue
}

type NewMetaOfType interface {
	NewMetaOf(*LState, LValue, LValue)
}

type MetaTableType interface {
	MetaTable(*LState, string) LValue
}

type MetaTableOfType interface {
	MetaTableOf(*LState, string) LValue
}

type FieldType interface {
	Field(string) string
}

type PackType interface {
	Unpack() any
}

type GenericType interface {
	LValue
	Unpack() any
	GobEncode() ([]byte, error)
	GobDecode(data []byte) error
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(bytes []byte) error
}

type Generic[T any] struct {
	Data T
	flag bool
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

func (gen *Generic[T]) String() string {
	var data any = gen.Data
	if v, ok := data.(fmt.Stringer); ok {
		return v.String()
	}

	return cast.ToString(gen.Data)
}
func (gen *Generic[T]) Type() LValueType { return LTGeneric }

func (gen *Generic[T]) AssertFloat64() (float64, bool) {
	var data any = gen.Data
	if v, ok := data.(Float64Type); ok {
		return v.AssertFloat64()
	}
	return 0, false
}

func (gen *Generic[T]) AssertString() (string, bool) {
	var data any = gen.Data
	if v, ok := data.(StringType); ok {
		return v.AssertString()
	}
	return "", false
}

func (gen *Generic[T]) AssertFunction() (*LFunction, bool) {
	var dat any = gen.Data
	switch vt := dat.(type) {
	case *LFunction:
		return vt, true
	case FunctionType:
		return vt.AssertFunction()
	case Call0:
		return NewCall0(vt), true
	case Call1:
		return NewCall1(vt), true
	case Call2:
		return NewCall2(vt), true
	case Call3:
		return NewCall3(vt), true
	case CallE0:
		return NewCallE0(vt), true
	case CallE1:
		return NewCallE1(vt), true
	case CallE2:
		return NewCallE2(vt), true
	case CallE3:
		return NewCallE3(vt), true
	default:
		return nil, false
	}
}

func (gen *Generic[T]) Hijack(fsm *CallFrameFSM) bool {
	var data any = gen.Data
	if v, ok := data.(HijackType); ok {
		return v.Hijack(fsm)
	}

	return false
}

func (gen *Generic[T]) IndexOf(L *LState, key string) LValue {
	if !gen.flag {
		return LNil
	}
	r := NewReflect(gen.Data)
	return r.IndexOf(L, key)
}

func (gen *Generic[T]) MetaOf(L *LState, key LValue) LValue {
	if !gen.flag {
		return LNil
	}
	r := NewReflect(gen.Data)
	return r.MetaOf(L, key)
}

func (gen *Generic[T]) Unpack() any {
	return gen.Data
}

func (gen *Generic[T]) Unwrap() T {
	return gen.Data
}

func NewGeneric[T any](data T) *Generic[T] {
	return &Generic[T]{
		Data: data,
	}
}

func NewGenericR[T any](data T) *Generic[T] {
	return &Generic[T]{
		Data: data,
		flag: true,
	}
}
