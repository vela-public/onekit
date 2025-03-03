package lua

import (
	"reflect"
	"time"
)

func NewCall0(fn Call0) *LFunction {
	return NewFunction(func(co *LState) int {
		fn()
		return 0
	})
}

func NewCall1(fn Call1) *LFunction {
	return NewFunction(func(co *LState) int {
		fn(co.Get(1))
		return 0
	})
}

func NewCall2(fn Call2) *LFunction {
	return NewFunction(func(co *LState) int {
		fn(co.Get(1), co.Get(2))
		return 0
	})
}

func NewCall3(fn Call3) *LFunction {
	return NewFunction(func(co *LState) int {
		fn(co.Get(1), co.Get(2), co.Get(3))
		return 0
	})
}

func NewCallE0(fn CallE0) *LFunction {
	return NewFunction(func(co *LState) int {
		if e := fn(); e != nil {
			co.Push(S2L(e.Error()))
			return 1
		}
		return 0
	})
}

func NewCallE1(fn CallE1) *LFunction {
	return NewFunction(func(co *LState) int {
		if e := fn(co.Get(1)); e != nil {
			co.Push(S2L(e.Error()))
			return 1
		}
		return 0
	})
}

func NewCallE2(fn CallE2) *LFunction {
	return NewFunction(func(co *LState) int {
		if e := fn(co.Get(1), co.Get(2)); e != nil {
			co.Push(S2L(e.Error()))
			return 1
		}
		return 0
	})
}

func NewCallE3(fn CallE3) *LFunction {
	return NewFunction(func(co *LState) int {
		if e := fn(co.Get(1), co.Get(2), co.Get(3)); e != nil {
			co.Push(S2L(e.Error()))
			return 1
		}
		return 0
	})
}

func Some(v any) LValue {
	lv, ok := TypeFor(v)
	if ok {
		return lv
	}
	return LNil
}

func TypeFor(v any) (LValue, bool) {
	switch vt := v.(type) {
	case nil:
		return LNil, true
	case LValue:
		return vt, true
	case bool:
		return LBool(vt), true
	case float64:
		return LNumber(vt), true
	case float32:
		return LNumber(vt), true
	case int8:
		return LInt(vt), true
	case int16:
		return LInt(vt), true
	case int32:
		return LNumber(vt), true
	case int:
		return LInt(vt), true
	case int64:
		return LNumber(vt), true

	case uint8:
		return LInt(vt), true
	case uint16:
		return LInt(vt), true
	case uint32:
		return LNumber(vt), true
	case uint:
		return LInt(vt), true
	case uint64:
		return LNumber(vt), true
	case []byte:
		return B2L(vt), true
	case string:
		return S2L(vt), true
	case time.Time:
		return Time(vt), true
	case []string:
		return SliceTo[string](vt), true
	case [][]byte:
		return SliceTo[[]byte](vt), true
	case []interface{}:
		return SliceTo[interface{}](vt), true
	case []int:
		return SliceTo[int](vt), true
	case []int8:
		return SliceTo[int8](vt), true
	case []int16:
		return SliceTo[int16](vt), true
	case []int32:
		return SliceTo[int32](vt), true
	case []int64:
		return SliceTo[int64](vt), true
	case []uint:
		return SliceTo[uint](vt), true
	case []uint16:
		return SliceTo[uint16](vt), true
	case []uint32:
		return SliceTo[uint32](vt), true
	case []uint64:
		return SliceTo[uint64](vt), true
	case []float32:
		return SliceTo[float32](vt), true
	case []float64:
		return SliceTo[float64](vt), true
	case []bool:
		return SliceTo[bool](vt), true
	case []LValue:
		return SliceTo[LValue](vt), true
	case map[string]string:
		return MapTo[string, string](vt), true
	case map[string]int:
		return MapTo[string, int](vt), true
	case map[string]int8:
		return MapTo[string, int8](vt), true
	case map[string]int16:
		return MapTo[string, int16](vt), true
	case map[string]int32:
		return MapTo[string, int32](vt), true
	case map[string]int64:
		return MapTo[string, int64](vt), true
	case map[string]uint:
		return MapTo[string, uint](vt), true
	case map[string]uint8:
		return MapTo[string, uint8](vt), true
	case map[string]uint16:
		return MapTo[string, uint16](vt), true
	case map[string]uint32:
		return MapTo[string, uint32](vt), true
	case map[string]uint64:
		return MapTo[string, uint64](vt), true
	case map[string]bool:
		return MapTo[string, bool](vt), true
	case map[string]float32:
		return MapTo[string, float32](vt), true
	case map[string]float64:
		return MapTo[string, float64](vt), true
	case map[string]LValue:
		return MapTo[string, LValue](vt), true
	case map[string]time.Time:
		return MapTo[string, time.Time](vt), true
	case map[string][]byte:
		return MapTo[string, []byte](vt), true
	case map[string]any:
		return MapTo[string, any](vt), true
	case map[int]string:
		return MapTo[int, string](vt), true
	case map[int][]byte:
		return MapTo[int, []byte](vt), true
	case map[int]int:
		return MapTo[int, int](vt), true
	case map[int]int8:
		return MapTo[int, int8](vt), true
	case map[int]int16:
		return MapTo[int, int16](vt), true
	case map[int]int32:
		return MapTo[int, int32](vt), true
	case map[int]int64:
		return MapTo[int, int64](vt), true
	case map[int]uint:
		return MapTo[int, uint](vt), true
	case map[int]uint8:
		return MapTo[int, uint8](vt), true
	case map[int]uint16:
		return MapTo[int, uint16](vt), true
	case map[int]uint32:
		return MapTo[int, uint32](vt), true
	case map[int]uint64:
		return MapTo[int, uint64](vt), true
	case map[int]bool:
		return MapTo[int, bool](vt), true
	case map[int]float32:
		return MapTo[int, float32](vt), true
	case map[int]float64:
		return MapTo[int, float64](vt), true
	case map[int]LValue:
		return MapTo[int, LValue](vt), true
	case map[int]time.Time:
		return MapTo[int, time.Time](vt), true
	case LGFunction:
		return NewFunction(vt), true
	case ValueType:
		return vt.ToValue(), true
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
	case reflect.Value:
		return NewReflect(vt).UnwrapLua(), true
	case error:
		if vt == nil {
			return LNil, true
		}
		return S2L(vt.Error()), true
	default:
		return LNil, false
	}
}

func ReflectTo[T any](t T) LValue {
	v, ok := TypeFor(t)
	if ok {
		return v
	}

	return NewReflect(t).UnwrapLua()
}
