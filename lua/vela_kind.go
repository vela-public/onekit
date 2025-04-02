package lua

func Kind(L *LState, v any, name string) LValue {
	switch kt := v.(type) {
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
	}

	return LNil
}
