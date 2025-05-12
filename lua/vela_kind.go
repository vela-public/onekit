package lua

func ValueOf(L *LState, v any, name string) LValue {
	switch kt := v.(type) {
	case map[string]string:
		return LString(kt[name])
	case IndexType:
		return kt.Index(L, name)
	case MetaType:
		return kt.Meta(L, LString(name))
	case MetaTableType:
		return kt.MetaTable(L, name)
	case UserKV:
		return kt.Get(name)
	case Getter:
		return ReflectTo(kt.Getter(name))
	case *LTable:
		return kt.RawGet(LString(name))
	}

	return LNil
}
