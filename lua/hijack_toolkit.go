package lua

func IF[T any](flag bool, a, b T) T {
	if flag {
		return a
	}
	return b
}

func Unwrap(v LValue) (LValue, any) {
	if v.Type() != LTGeneric {
		return v, v
	}
	return v, v.(GenericType).Unpack()
}
