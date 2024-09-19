package luakit

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
)

var (
	overflowE = errors.New("Index over flow")
	tooSmallE = errors.New("Index too small")
	invalidE  = errors.New("invalid slice value")
)

type Slice[T any] []T

func (s Slice[T]) String() string {
	text, err := json.Marshal(s)
	if err != nil {
		return "[]"
	}
	return B2S(text)
}

func (s Slice[T]) Type() lua.LValueType                   { return lua.LTSlice }
func (s Slice[T]) AssertFloat64() (float64, bool)         { return float64(len(s)), true }
func (s Slice[T]) AssertString() (string, bool)           { return "", false }
func (s Slice[T]) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (s Slice[T]) Hijack(fsm *lua.CallFrameFSM) bool      { return false }
func (s Slice[T]) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "size", "sz":
		return lua.LInt(len(s))
	case "concat":
		return lua.NewFunction(s.concatL)
	default:
		return lua.LNil
	}
}

func (s Slice[T]) Meta(L *lua.LState, key lua.LValue) lua.LValue {
	i, ok := key.AssertFloat64()
	if !ok {
		return lua.LNil
	}

	var idx int

	if i < 0 {
		idx = s.Len() + int(i)
	} else {
		idx = int(i) - 1
	}

	val, ok := s.Get(idx)
	if !ok {
		return lua.LNil
	}
	return lua.ToLValue(val)
}

func (s Slice[T]) NewMeta(L *lua.LState, key lua.LValue, val lua.LValue) {
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

	_ = s.Set(idx, val)
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

func (s *Slice[T]) Set(idx int, val lua.LValue) error {
	a := *s
	n := len(a)
	if idx < 0 {
		return overflowE
	}

	if idx < n {
		if v, ok := val.(T); ok {
			a[idx] = v
			*s = a
			return nil
		}
		return invalidE
	}

	switch val.Type() {
	case lua.LTNil:
		return invalidE
	default:
		if v, ok := val.(T); ok {
			a = append(a, v)
			*s = a
			return nil
		}
		return invalidE
	}
}

// concatL is a lua function that concatenates the slice
func (s *Slice[T]) concatL(L *lua.LState) int {
	if s.Len() == 0 {
		L.Push(lua.EmptyString)
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

func NewSliceL(L *lua.LState) int { // luakit.slice{"123" , "123"}
	n := L.GetTop()
	if n == 0 {
		L.Push(Slice[lua.LValue]{})
		return 1
	}

	s := make(Slice[lua.LValue], n)
	for i := 1; i <= n; i++ {
		s[i-1] = L.Get(i)
	}

	L.Push(s)
	return 1
}
