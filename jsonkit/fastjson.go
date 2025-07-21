package jsonkit

import (
	"fmt"
	"github.com/valyala/fastjson"
	"github.com/valyala/fastjson/fastfloat"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/todo"
	"sort"
	"strconv"
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

func (f *FastJSON) Wrap() *fastjson.Value {
	return f.value
}

func (f *FastJSON) Get(key string) *fastjson.Value {
	value := f.cache.Get(key)
	if value != nil {
		return value
	}

	value = todo.Or[fastjson.Value](f.value.Get(key), Empty)
	f.cache.Set(key, value)
	return value
}

func (f *FastJSON) Object() (*fastjson.Object, error) {
	obj, err := f.value.Object()
	return obj, err
}

func (f *FastJSON) setText(key string, v string) error {
	sz := len(v)

	var text string
	if sz >= 2 && v[0] == '"' && v[sz-1] == '"' {
		text = v
		return nil
	} else {
		text = strconv.Quote(v)
	}

	dat, err := fastjson.Parse(text)
	if err != nil {
		return err
	}
	f.value.Set(key, dat)
	f.cache.Set(key, dat)
	return nil
}

func (f *FastJSON) Settle(key string, v any) error {
	obj, err := f.Object()
	if err != nil {
		return err
	}

	switch dat := v.(type) {
	case string:
		return f.setText(key, dat)
	case []byte:
		return f.setText(key, cast.B2S(dat))
	case *fastjson.Value:
		obj.Set(key, dat)
		f.cache.Set(key, dat)
		return nil
	case bool:
		val := fastjson.MustParse(todo.IF(dat, "true", "false"))
		obj.Set(key, val)
		f.cache.Set(key, val)
		return nil
	case int, int32, int64, uint, uint32, uint64, float32, float64:
		val := fastjson.MustParse(cast.ToString(dat))
		obj.Set(key, val)
		f.cache.Set(key, val)
		return nil
	case nil:
		obj.Set(key, Empty)
		f.cache.Set(key, Empty)
		return nil
	default:
		return fmt.Errorf("invalid type")
	}
}

func (f *FastJSON) Clean(keys ...string) (changed bool) {
	if len(keys) == 0 {
		return
	}

	obj, err := f.Object()
	if err != nil {
		return
	}

	sz := obj.Len()
	for _, key := range keys {
		obj.Del(key)
	}

	if obj.Len() == sz {
		return false
	}
	return true
}

func (f *FastJSON) Delete(dat ...string) (changed bool) {
	obj, err := f.Object()
	if err != nil {
		return
	}

	sort.Strings(dat)
	sz := len(dat)

	obj.Visit(func(key []byte, v *fastjson.Value) {
		k := cast.B2S(key)
		if i := sort.SearchStrings(dat, k); i < sz && dat[i] == k {
			*v = fastjson.Value{}
			f.cache.Set(k, Empty)
			changed = true
		}
	})
	return
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

func (f *FastJSON) Chunk(key string) []byte {
	return cast.S2B(f.Text(key))
}

func (f *FastJSON) Text(key string) string {
	value := f.Get(key)
	if value.Type() == fastjson.TypeNull {
		return ""
	}

	return Unquote(value.String())
}

func (f *FastJSON) To(key string) any {
	v := f.Get(key)
	switch v.Type() {
	case fastjson.TypeNull:
		return nil

	case fastjson.TypeString:
		return Unquote(v.String())

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
		return Unquote(v.String())
	}

}

func (f *FastJSON) ParseText(body string) todo.Result[*FastJSON, error] {
	if len(body) == 0 {
		f.value = Empty
		return todo.Err[*FastJSON, error](f, fmt.Errorf("empty body"))
	}

	v, err := fastjson.Parse(body)
	if err != nil {
		f.value = Empty
		return todo.Err(f, err)
	}

	f.value = v
	return todo.NewOK[*FastJSON, error](f)
}

func (f *FastJSON) Parse(p *fastjson.Parser, body string) todo.Result[*FastJSON, error] {
	v, err := p.Parse(body)
	if err != nil {
		f.value = Empty
		return todo.Err(f, err)
	}

	f.value = v
	return todo.NewOK[*FastJSON, error](f)
}

func (f *FastJSON) visit(key string) lua.LValue {
	v := f.Get(key)

	switch v.Type() {
	case fastjson.TypeNull:
		return lua.LNil

	case fastjson.TypeString:
		item := v.String()
		return lua.S2L(Unquote(item))

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
		return lua.S2L(Unquote(v.String())) //typeRawString 7
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
		f.value.Set(key, fastjson.MustParse(Quote(val.String())))
	default:
		f.value.Set(key, fastjson.MustParse(Quote(val.String())))
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
		f.value.Set(key, fastjson.MustParse(Quote(val.String())))
	default:
		f.value.Set(key, fastjson.MustParse(Quote(val.String())))
	}
}
