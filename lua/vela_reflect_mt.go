package lua

import (
	"errors"
	"fmt"
	"reflect"
)

type Reflect[T any] struct {
	t     reflect.Type
	v     reflect.Value
	inner struct {
		d T             // 原始数据
		t reflect.Type  //原始数据类型
		v reflect.Value //原始数据值
	}
}

func (r *Reflect[T]) String() string                     { return fmt.Sprintf("reflect<%T>", r.inner.d) }
func (r *Reflect[T]) Type() LValueType                   { return LTObject }
func (r *Reflect[T]) AssertFloat64() (float64, bool)     { return 0, false }
func (r *Reflect[T]) AssertString() (string, bool)       { return "", false }
func (r *Reflect[T]) AssertFunction() (*LFunction, bool) { return NewFunction(r.Call), true }
func (r *Reflect[T]) Hijack(fsm *CallFrameFSM) bool      { return false }

func (r *Reflect[T]) UnwrapData() any {
	return r.inner.d
}

func (r *Reflect[T]) Redirect() {
	vt := r.t
	vo := r.v

	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
		vo = vo.Elem()

		r.t = vt
		r.v = vo
		r.Redirect()
	}
}

func (r *Reflect[T]) NoKey() bool {
	vt := r.t
	switch vt.Kind() {
	case reflect.Map:
		return false
	case reflect.Struct:
		return false
	default:
		return true
	}
}

func (r *Reflect[T]) UnwrapLua() LValue {
	vt := r.t
	vo := r.v
	switch vt.Kind() {
	case reflect.Invalid:
		return LNil
	case reflect.Bool:
		return LBool(vo.Bool())
	case reflect.Int:
		return LInt(vo.Int())
	case reflect.Int8:
		return LInt(vo.Int())
	case reflect.Int16:
		return LInt(vo.Int())
	case reflect.Int32:
		return LInt(vo.Int())
	case reflect.Int64:
		return LInt64(vo.Int())
	case reflect.Uint:
		return LUint(vo.Uint())
	case reflect.Uint8:
		return LUint(vo.Uint())
	case reflect.Uint16:
		return LUint(vo.Uint())
	case reflect.Uint32:
		return LUint(vo.Uint())
	case reflect.Uint64:
		return LUint64(vo.Uint())
	//case reflect.Uintptr:

	case reflect.Float32:
		return LNumber(vo.Float())
	case reflect.Float64:
		return LNumber(vo.Float())
	case reflect.Array:
		return r
	case reflect.Chan:
		return r
	case reflect.Func:
		return r
	case reflect.Interface:
		return r
	case reflect.Map:
		return r
	case reflect.Pointer:
		return r
	case reflect.Slice:
		return r
	case reflect.String:
		return S2L(vo.String())
	case reflect.Struct:
		return r
		//case reflect.UnsafePointer:
	default:
		return LNil

	}
}

func (r *Reflect[T]) CallFunc(args ...interface{}) ([]reflect.Value, error) {

	if r.v.IsZero() {
		return nil, errors.New("fn must not be empty")
	}

	if !r.t.IsVariadic() && len(args) != r.t.NumIn() {
		return nil, fmt.Errorf("fn params num is %d, but got %d", r.t.NumIn(), len(args))
	}

	if r.t.IsVariadic() && len(args) < r.t.NumIn()-1 {
		return nil, fmt.Errorf("fn params num is %d at least, but got %d", r.t.NumIn()-1, len(args))
	}

	a := make([]reflect.Value, len(args))
	for i, v := range args {
		if v == nil {
			if r.t.IsVariadic() && i >= r.t.NumIn()-1 {
				a[i] = reflect.New(r.t.In(r.t.NumIn() - 1).Elem()).Elem()
			} else {
				a[i] = reflect.New(r.t.In(i)).Elem()
			}
		} else {
			a[i] = reflect.ValueOf(v)
		}
	}

	rc := r.v.Call(a)
	return rc, nil
}

func (r *Reflect[T]) Call(L *LState) int {
	if r.t.Kind() != reflect.Func {
		L.RaiseError("%s not function", r.t.Name())
		return 0
	}

	args := UnpackGo(L)
	rc, err := r.CallFunc(args...)
	if err != nil {
		L.Push(LNil)
		L.Push(S2L(err.Error()))
		return 2
	}

	L.Push(SliceTo[reflect.Value](rc))
	return 1
}

func (r *Reflect[T]) Index(L *LState, key string) (lv LValue) {

	switch d := any(r.inner.d).(type) {
	case IndexType:
		lv = d.Index(L, key)
	case Getter:
		v := d.Getter(key)
		if v != nil {
			lv = NewReflect(v).UnwrapLua()
		}
	}

	if lv != nil && lv.Type() != LTNil {
		return lv
	}

	v, ok := r.Filed(key)
	if ok {
		return NewReflect(v).UnwrapLua()
	}

	m, ok := r.Method(key)
	if ok {
		return NewReflect(m)
	}

	return LNil
}

func (r *Reflect[T]) Method(key string) (reflect.Value, bool) {
	n := r.inner.t.NumMethod()
	for i := 0; i < n; i++ {
		m := r.inner.t.Method(i)
		if m.Name == key {
			return r.inner.v.Method(i), true
		}
	}
	return reflect.Value{}, false
}

func (r *Reflect[T]) Filed(key string) (reflect.Value, bool) {
	var rc reflect.Value
	if r.NoKey() {
		return rc, false
	}

	switch r.t.Kind() {
	case reflect.Map:
		return r.v.MapIndex(reflect.ValueOf(key)), true
	case reflect.Struct:
		n := r.v.NumField()
		for i := 0; i < n; i++ {
			f := r.t.Field(i)
			if f.Tag.Get("lua") == key {
				return r.v.Field(i), true
			}
			if f.Name == key {
				return r.v.Field(i), true
			}
		}
		return rc, false
	default:
		return rc, false
	}

}

func NewReflect[T any](t T) *Reflect[T] {
	r := &Reflect[T]{}
	r.inner.d = t

	switch dat := any(t).(type) {
	case reflect.Value:
		r.inner.t = dat.Type()
		r.inner.v = dat

	default:
		r.inner.t = reflect.TypeOf(t)
		r.inner.v = reflect.ValueOf(t)
	}

	r.t = r.inner.t
	r.v = r.inner.v

	r.Redirect()
	return r
}
