package abi

import (
	"encoding/json"
	"github.com/vela-public/onekit/lua"
	"unsafe"
)

type Int32 struct {
	value []int32
}

func (i32 *Int32) UnsafeText() []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&i32.value[0])), i32.Length()*4)
}

func (i32 *Int32) Text() []byte {
	text, _ := json.Marshal(i32.value)
	return text
}

func (i32 *Int32) String() string                         { return string(i32.UnsafeText()) }
func (i32 *Int32) Type() lua.LValueType                   { return lua.LTObject }
func (i32 *Int32) AssertFloat64() (float64, bool)         { return float64(i32.Length()), true }
func (i32 *Int32) AssertString() (string, bool)           { return string(i32.Text()), true }
func (i32 *Int32) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (i32 *Int32) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (i32 *Int32) Length() int {
	return len(i32.value)
}

func (i32 *Int32) Get(n int) (int32, bool) {
	sz := i32.Length()
	if n < 0 || n >= sz {
		return 0, false
	}

	return i32.value[n], true
}

func (i32 *Int32) Set(n int, v int32) bool {
	sz := i32.Length()
	if n < 0 || n >= sz {
		return false
	}

	i32.value[n] = v
	return true
}

func (i32 *Int32) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "len":
		return lua.LInt(i32.Length())
	case "bin":
		return lua.B2L(i32.UnsafeText())
	case "text":
		return lua.B2L(i32.Text())
	default:
		return lua.LNil
	}
}

func (i32 *Int32) NewIndex(L *lua.LState, key string, val lua.LValue) {
	return
}

func (i32 *Int32) Meta(L *lua.LState, key lua.LValue) lua.LValue {
	idx, ok := lua.Must[int](key)
	if !ok {
		return lua.LNil
	}
	y, ok := i32.Get(idx)
	if ok {
		return lua.LInt(y)
	}
	return lua.LNil
}

func (i32 *Int32) NewMeta(L *lua.LState, key lua.LValue, val lua.LValue) {
	v, ok := lua.Must[int32](val)
	if !ok {
		return
	}

	idx, ok := lua.Must[int](key)
	if !ok {
		return
	}
	i32.Set(idx, v)
}

func NewInt32(n int) *Int32 {
	return &Int32{
		value: make([]int32, n),
	}
}

func NewInt32L(L *lua.LState) int {
	val := L.Get(1)
	n, ok := lua.Must[int](L.Get(1))
	if !ok {
		L.RaiseError("cap must be number got %s", val.Type().String())
		return 0
	}
	i32 := NewInt32(n)
	L.Push(i32)
	return 1
}
