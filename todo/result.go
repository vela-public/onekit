package todo

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
)

const (
	Undefined ResultType = iota
	Succeed
	Invalid
)

type ResultType uint8

// Result 模拟 Rust 的 Result 类型
type Result[T, E any] struct {
	Value T
	Error E
	Flag  ResultType
}

func Ok[T, E any](v T) Result[T, E] {
	return Result[T, E]{
		Value: v,
		Flag:  Succeed,
	}
}

func NewE[T, E any](e E) Result[T, E] {
	return Result[T, E]{
		Error: e,
		Flag:  Invalid,
	}
}

func (r Result[T, E]) String() string {
	switch r.Flag {
	case Succeed:
		return cast.ToString(r.Value)
	case Undefined:
		return ""
	case Invalid:
		return ""
	}

	return ""
}

func (r Result[T, E]) Bad() bool {
	return r.Flag == Invalid || r.Flag == Undefined
}

func (r Result[T, E]) Type() lua.LValueType {
	return lua.LTObject
}

func (r Result[T, E]) AssertFloat64() (float64, bool) {
	switch v := any(r.Value).(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case lua.LInt:
		return float64(v), true
	case lua.LInt64:
		return float64(v), true
	case lua.LNumber:
		return float64(v), true
	case lua.LUint:
		return float64(v), true
	case lua.LUint64:
		return float64(v), true
	default:
		return 0, false
	}
}

func (r Result[T, E]) AssertString() (string, bool) {
	var data any
	switch r.Flag {
	case Succeed:
		data = r.Value
	case Invalid:
		return "", false
	case Undefined:
		return "", false
	default:
		return "", false

	}

	switch v := data.(type) {
	case string:
		return v, true
	case []byte:
		return cast.B2S(v), true
	case lua.LString:
		return string(v), true
	case fmt.Stringer:
		return v.String(), true
	case error:
		return v.Error(), true
	default:
		return "", false
	}
}

func (r Result[T, E]) AssertFunction() (*lua.LFunction, bool) {
	var data any
	switch r.Flag {
	case Succeed:
		data = r.Value
	case Invalid:
		data = r.Error
		return nil, false
	case Undefined:
		return nil, false
	}

	switch v := data.(type) {
	case *lua.LFunction:
		return v, true
	case lua.FunctionType:
		return v.AssertFunction()
	default:
		return nil, false
	}
}

func (r Result[T, E]) Hijack(fsm *lua.CallFrameFSM) bool {

	if fsm.OpCode() == lua.OP_TEST {
		fsm.LVAsBool(r.Flag == Succeed)
		return true
	}

	var data any
	switch r.Flag {
	case Undefined:
		return false
	case Invalid:
		data = r.Error
	case Succeed:
		data = r.Value
	}

	v, ok := data.(interface{ Hijack(*lua.CallFrameFSM) bool })
	if ok {
		return v.Hijack(fsm)
	}
	return false
}

func NewOK[T, E any](value T) Result[T, E] {
	return Result[T, E]{Value: value, Flag: Succeed}
}

func Err[T, E any](v T, err E) Result[T, E] {
	return Result[T, E]{Value: v, Error: err, Flag: Invalid}
}

func (r Result[T, E]) Unwrap() (t T, ok bool) { return r.Value, r.Flag == Succeed }
func (r Result[T, E]) UnwrapErr() E           { return r.Error }

func (r Result[T, E]) Used() bool {
	if r.Flag == Undefined {
		return false
	}
	return true
}

func (r Result[T, E]) Expect(fn func(E)) {
	if r.Flag == Invalid {
		fn(r.Error)
	}
}

func (r Result[T, E]) Y(fn func(T)) Result[T, E] {
	if r.Flag == Succeed {
		fn(r.Value)
	}
	return r
}

func (r Result[T, E]) N(fn func(E)) Result[T, E] {
	if r.Flag == Invalid {
		fn(r.Error)
	}
	return r
}

func (r Result[T, E]) Index(L *lua.LState, key string) lua.LValue {
	if r.Flag != Succeed {
		return lua.LNil
	}

	x, ok := any(r.Value).(lua.IndexType)
	if ok {
		return x.Index(L, key)
	}
	return lua.LNil
}

func (r Result[T, E]) NewIndex(L *lua.LState, key string, val lua.LValue) {
	if r.Flag != Succeed {
		return
	}

	x, ok := any(r.Value).(lua.NewIndexType)
	if !ok {
		return
	}

	x.NewIndex(L, key, val)
}
