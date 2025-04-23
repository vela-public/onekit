package lua

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/vela-public/onekit/cast"
)

const (
	SliceLibName = "slice"
)

var (
	OverflowE = errors.New("Index over flow")
	TooSmallE = errors.New("Index too small")
	InvalidE  = errors.New("invalid slice value")
)

type Slice[T any] []T

func (s *Slice[T]) String() string                     { return cast.B2S(s.Bytes()) }
func (s *Slice[T]) Type() LValueType                   { return LTSlice }
func (s *Slice[T]) AssertFloat64() (float64, bool)     { return float64(s.Len()), true }
func (s *Slice[T]) AssertString() (string, bool)       { return "", false }
func (s *Slice[T]) AssertFunction() (*LFunction, bool) { return nil, false }
func (s *Slice[T]) Hijack(fsm *CallFrameFSM) bool      { return false }

func (s *Slice[T]) Bytes() []byte {
	text, err := s.MarshalJSON()
	if err != nil {
		return []byte("[]")
	}
	return text
}

func (s *Slice[T]) Append(v T) {
	*s = append(*s, v)
}

func (s *Slice[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(*s)
}
func (s *Slice[T]) Index(L *LState, key string) LValue {
	switch key {
	case "size", "sz":
		return LInt(s.Len())
	case "concat":
		return NewFunction(s.concatL)
	default:
		return LNil
	}
}

func (s *Slice[T]) Meta(L *LState, key LValue) LValue {
	i, ok := key.AssertFloat64()
	if !ok {
		return LNil
	}

	var idx int

	if i < 0 {
		idx = s.Len() + int(i)
	} else {
		idx = int(i) - 1
	}

	val, ok := s.Get(idx)
	if !ok {
		return LNil
	}
	return ReflectTo(val)
}

func (s *Slice[T]) NewMeta(L *LState, key LValue, val LValue) {
	i, ok := key.AssertFloat64()
	if !ok {
		return
	}

	var idx int
	if i < 0 {
		idx = s.Len() + int(i)
	} else {
		idx = int(i) - 1
	}

	_ = s.Set(idx, Check[T](L, val))
}

func (s *Slice[T]) Len() int {
	return len(*s)
}

func (s *Slice[T]) Get(idx int) (t T, ok bool) {
	a := *s

	if idx < 0 {
		ok = false
		return
	}

	if idx >= len(a) {
		ok = false
		return
	}

	return a[idx], true
}

func (s *Slice[T]) Set(idx int, v T) error {
	a := *s
	n := len(a)
	if idx < 0 {
		return TooSmallE
	}

	if idx < n {
		a[idx] = v
		*s = a
		return nil
	}

	if idx == n {
		a = append(a, v)
		*s = a
		return nil
	}

	return OverflowE
}

// concatL is a lua function that concatenates the slice
func (s *Slice[T]) concatL(L *LState) int {
	if s.Len() == 0 {
		L.Push(EmptyString)
		return 1
	}

	sep := L.CheckString(1)
	v := *s

	var buf bytes.Buffer
	for i, item := range v {
		if i > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(cast.ToString(item))
	}

	L.Push(S2L(buf.String()))
	return 1
}

func NewSliceStringL(L *LState) int {
	top := L.GetTop()
	s := NewSlice[string](top)
	for i := 1; i <= top; i++ {
		_ = s.Set(i-1, Check[string](L, L.Get(i)))
	}
	L.Push(s)
	return 1
}

func NewSliceIntL(L *LState) int {
	top := L.GetTop()
	s := NewSlice[int](top)
	for i := 1; i <= top; i++ {
		_ = s.Set(i-1, Check[int](L, L.Get(i)))
	}
	L.Push(s)
	return 1
}

func NewSliceBoolL(L *LState) int {
	top := L.GetTop()
	s := NewSlice[bool](top)
	for i := 1; i <= top; i++ {
		_ = s.Set(i-1, Check[bool](L, L.Get(i)))
	}
	L.Push(s)
	return 1
}

func NewSliceFloatL(L *LState) int {
	top := L.GetTop()
	s := NewSlice[float64](top)
	for i := 1; i <= top; i++ {
		_ = s.Set(i-1, Check[float64](L, L.Get(i)))
	}
	L.Push(s)
	return 1
}

func NewSlice[T any](cap int) *Slice[T] { // luakit.slice{"123", "123"}
	if cap == 0 {
		return &Slice[T]{}
	}

	s := make(Slice[T], cap)
	return &s
}

func SliceTo[T any](arr []T) *Slice[T] {
	s := new(Slice[T])
	*s = arr
	return s
}

func NewSliceL(L *LState) int { // luakit.slice{"123" , "123"}
	n := L.GetTop()
	s := NewSlice[LValue](n)
	if n == 0 {
		L.Push(s)
		return 1
	}

	for i := 1; i <= n; i++ {
		_ = s.Set(i-1, L.Get(i))
	}
	L.Push(s)
	return 1
}

func SliceIndex(L *LState, key string) LValue {
	switch key {
	case "s":
		return NewFunction(NewSliceStringL)
	case "d":
		return NewFunction(NewSliceIntL)
	case "f":
		return NewFunction(NewSliceFloatL)
	case "b":
		return NewFunction(NewSliceBoolL)
	}

	return LNil
}

/*
	local s = slice.s("v" , "e" , 'l' , 'a')
	local n = slice.d(1,12)
	local b = slice.b(true, false)
	print(s.sz)
	print(s[1])
	print(n[2])
	print(b[2]
*/

func OpenSliceLib(L *LState) int {
	mod := NewExport("lua.slice.export", WithFunc(NewSliceL), WithIndex(SliceIndex))
	L.SetGlobal("slice", mod)
	L.Push(mod)
	return 1
}
