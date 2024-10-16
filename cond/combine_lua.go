package cond

import (
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/jsonkit"
	"github.com/vela-public/onekit/lua"
)

func (iv *Combine) String() string                         { return cast.B2S(iv.Json()) }
func (iv *Combine) Type() lua.LValueType                   { return lua.LTObject }
func (iv *Combine) AssertFloat64() (float64, bool)         { return 0, false }
func (iv *Combine) AssertString() (string, bool)           { return "", false }
func (iv *Combine) AssertFunction() (*lua.LFunction, bool) { return lua.NewFunction(iv.Call), true }
func (iv *Combine) Hijack(*lua.CallFrameFSM) bool          { return false }

func (iv *Combine) Json() []byte {
	v := *iv
	n := len(v)
	if n == 0 {
		return []byte("[]")
	}

	enc := jsonkit.NewJson()
	enc.Arr("")
	for i := 0; i < n; i++ {
		enc.Val(v[i].String())
		enc.Val(",")
	}
	enc.End("]")
	return enc.Bytes()
}

func (iv *Combine) cndL(L *lua.LState) int {
	iv.CheckMany(L, LState(L))
	L.Push(iv)
	return 1
}

func (iv *Combine) MatchL(L *lua.LState) int {
	obj := L.Get(1)
	L.Push(lua.LBool(iv.Match(obj, LState(L))))
	return 1
}

func (iv *Combine) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "cnd":
		return lua.NewFunction(iv.cndL)
	case "match":
		return lua.NewFunction(iv.MatchL)
	default:
		return lua.LNil
	}
}

func (iv *Combine) Call(L *lua.LState) int {
	ret := iv.Match(L.Get(1), LState(L))
	L.Push(lua.LBool(ret))
	return 1
}

func NewCombineL(L *lua.LState) int {
	c := NewCombine()
	L.Push(c)
	return 1
}
