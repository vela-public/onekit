package lua

func ValueOf(L *LState, v any, name string) LValue {
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
