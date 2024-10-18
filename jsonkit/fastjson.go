package jsonkit

import (
	"github.com/valyala/fastjson"
	"github.com/valyala/fastjson/fastfloat"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/todo"
)

var Empty = &fastjson.Value{}

type FastJSON struct {
	value *fastjson.Value
	cache libkit.DataKV[string, *fastjson.Value]
}

func (f *FastJSON) String() string                         { return f.value.String() }
func (f *FastJSON) Type() lua.LValueType                   { return lua.LTObject }
func (f *FastJSON) AssertFloat64() (float64, bool)         { return 0, false }
func (f *FastJSON) AssertString() (string, bool)           { return "", false }
func (f *FastJSON) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (f *FastJSON) Hijack(*lua.CallFrameFSM) bool          { return false }

func (f *FastJSON) Get(key string) *fastjson.Value {
	value := f.cache.Get(key)
	if value != nil {
		return value
	}

	value = todo.Or[fastjson.Value](f.value.Get(key), Empty)
	f.cache.Set(key, value)
	return value
}

func (f *FastJSON) Int(key string) int {
	n, err := f.Get(key).Int()
	if err != nil {
		return 0
	}
	return n
}

func (f *FastJSON) Int64(key string) int64 {
	n, err := f.Get(key).Int64()
	if err != nil {
		return 0
	}
	return n
}

func (f *FastJSON) Uint(key string) uint {
	n, err := f.Get(key).Uint()
	if err != nil {
		return 0
	}
	return n
}

func (f *FastJSON) Uint64(key string) uint64 {
	n, err := f.Get(key).Uint64()
	if err != nil {
		return 0
	}
	return n
}

func (f *FastJSON) Float64(key string) float64 {
	n, err := f.Get(key).Float64()
	if err != nil {
		return 0
	}
	return n
}

func (f *FastJSON) Bool(key string) bool {
	b, err := f.Get(key).Bool()
	if err != nil {
		return false
	}
	return b
}

func (f *FastJSON) Text(key string) string {
	data := f.Get(key)
	text := data.MarshalTo(nil)

	if data.Type() == fastjson.TypeString && len(text) >= 2 && text[0] == '"' {
		return cast.B2S(text[1 : len(text)-1])
	}
	return cast.B2S(text)
}

func (f *FastJSON) Unwrap(key string) any {
	v := f.Get(key)
	switch v.Type() {
	case fastjson.TypeNull:
		return nil

	case fastjson.TypeString:
		item := v.String()
		return item[1 : len(item)-1]

	case fastjson.TypeNumber:
		n, err := fastfloat.Parse(v.String())
		if err != nil {
			return float64(0)
		}
		return n

	case fastjson.TypeObject:
		return &FastJSON{value: v}

	case fastjson.TypeArray:
		return &FastJSON{value: v}

	case fastjson.TypeTrue:
		return true

	case fastjson.TypeFalse:
		return false

	default:
		return v.String()
	}

}

func (f *FastJSON) Parse(body string) todo.Result[*FastJSON, error] {
	v, err := fastjson.Parse(body)
	if err != nil {
		f.value = Empty
		return todo.Err(f, err)
	}

	f.value = v
	return todo.NewOK[*FastJSON, error](f)
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

func (f *FastJSON) NewMeta(L *lua.LState, k lua.LValue, val lua.LValue) {
	key := k.String()

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
