package jsonkit

import (
	"github.com/valyala/fastjson"
	"github.com/valyala/fastjson/fastfloat"
	"github.com/vela-public/onekit/lua"
)

type FastJSON struct {
	value *fastjson.Value
}

func (f *FastJSON) String() string                         { return f.value.String() }
func (f *FastJSON) Type() lua.LValueType                   { return lua.LTObject }
func (f *FastJSON) AssertFloat64() (float64, bool)         { return 0, false }
func (f *FastJSON) AssertString() (string, bool)           { return "", false }
func (f *FastJSON) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (f *FastJSON) Hijack(*lua.CallFrameFSM) bool          { return false }

func (f *FastJSON) Int(L *lua.LState) int {
	key := L.CheckString(1)
	n := f.value.GetInt(key)
	L.Push(lua.LNumber(n))
	return 1
}

func (f *FastJSON) Str(L *lua.LState) int {
	key := L.CheckString(1)
	b := f.value.GetStringBytes(key)
	L.Push(lua.LString(lua.B2S(b)))
	return 1
}

func (f *FastJSON) Bool(L *lua.LState) int {
	key := L.CheckString(1)
	b := f.value.GetBool(key)
	L.Push(lua.LBool(b))
	return 1
}

func (f *FastJSON) ParseBytes(body []byte) error {
	v, err := fastjson.ParseBytes(body)
	if err != nil {
		return err
	}
	f.value = v

	return nil
}

func (f *FastJSON) Parse(body string) error {
	v, err := fastjson.Parse(body)
	if err != nil {
		return err
	}
	f.value = v
	return nil
}

func (f *FastJSON) visit(key string) lua.LValue {
	if f.value == nil {
		return lua.LNil
	}

	v := f.value.Get(key)
	if v == nil {
		return lua.LNil
	}

	switch v.Type() {
	case fastjson.TypeNull:
		return lua.LNil

	case fastjson.TypeString:
		item := v.String()
		return lua.S2L(item[1 : len(item)-1])

	case fastjson.TypeNumber:
		n, err := fastfloat.Parse(v.String())
		if err != nil {
			return lua.LNil
		}
		return lua.LNumber(n)

	case fastjson.TypeObject:
		return &FastJSON{value: v}

	case fastjson.TypeArray:
		return &FastJSON{value: v}

	case fastjson.TypeTrue:
		return lua.LTrue

	case fastjson.TypeFalse:
		return lua.LFalse

	default:
		return lua.S2L(v.String()) //typeRawString 7
	}

}

func (f *FastJSON) Index(L *lua.LState, key string) lua.LValue {
	return f.visit(key)
}

func (f *FastJSON) Meta(L *lua.LState, key lua.LValue) lua.LValue {
	return f.visit(key.String())
}

func (f *FastJSON) NewIndex(L *lua.LState, key string, val lua.LValue) {
	switch val.Type() {
	case lua.LTNil:
		return
	case lua.LTNumber, lua.LTBool, lua.LTInt, lua.LTInt64, lua.LTUint, lua.LTUint64:
		fv, err := fastjson.Parse(val.String())
		if err != nil {
			L.RaiseError("fastjson decode fail %v", err)
			return
		}

		f.value.Set(key, fv)
	case lua.LTObject:
		item, ok := val.(*FastJSON)
		if ok {
			f.value.Set(key, item.value)
			return
		}

		fv, err := fastjson.Parse(val.String())
		if err != nil {
			L.RaiseError("fastjson decode fail %v", err)
			return
		}

		f.value.Set(key, fv)
	case lua.LTString:
		f.value.Set(key, fastjson.MustParse("\""+val.String()+"\""))
	default:
		f.value.Set(key, fastjson.MustParse("\""+val.String()+"\""))
	}
}
