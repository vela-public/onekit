package lua

import (
	"bytes"
	"strings"
)

func StringOr(lv LValue, s string) string {
	if lv == nil {
		return s
	}
	if lv.Type() == LTNil {
		return s
	}

	txt := lv.String()
	if txt == "" {
		return s
	}

	return txt
}

func StringsOr(L *LState, v any, key string, placeholders string) string {
	keys := strings.Split(key, "-")
	sz := len(keys)
	if sz == 1 {
		return StringOr(ValueOf(L, v, key), placeholders)
	}

	var buff bytes.Buffer

	for i := 0; i < sz; i++ {
		k := keys[i]
		if i != 0 {
			buff.WriteString("-")
		}
		buff.WriteString(StringOr(ValueOf(L, v, k), placeholders))
	}

	return buff.String()
}

func ValueOf(L *LState, v any, name string) LValue {
	if v == nil {
		return LNil
	}

	switch kt := v.(type) {
	case map[any]any:
		return ReflectTo(kt[name])
	case map[string]string:
		return LString(kt[name])
	case map[string]int:
		return LNumber(kt[name])
	case map[string]float64:
		return LNumber(kt[name])
	case map[string]bool:
		return LBool(kt[name])
	case map[string]any:
		return ReflectTo(kt[name])
	case map[string]LValue:
		return kt[name]
	case IndexType:
		return kt.Index(L, name)
	case MetaType:
		return kt.Meta(L, LString(name))
	case MetaTableType:
		return kt.MetaTable(L, name)
	case UserKV:
		return kt.Get(name)
	case safeUserKV:
		return kt.Get(name)
	case Getter:
		return ReflectTo(kt.Getter(name))
	case *LTable:
		return kt.RawGet(LString(name))
	case IndexOfType:
		return kt.IndexOf(L, name)
	case MetaOfType:
		return kt.MetaOf(L, LString(name))
	case MetaTableOfType:
		return kt.MetaTableOf(L, name)
	case LUserData:
		return ValueOf(L, kt.Value, name)
	case PackType:
		return ValueOf(L, kt.Unpack(), name)
	case func(string) string:
		return LString(kt(name))
	default:
		r := NewReflect(v)
		return r.IndexOf(L, name)
	}
}
