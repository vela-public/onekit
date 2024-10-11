package todo

func IF[T any](flag bool, a, b T) T {
	if flag {
		return a
	}
	return b
}

func Fn[T any](fn func() error, a, b T) T {
	if err := fn(); err != nil {
		return b
	}
	return a
}

func Or[T any](a *T, b *T) *T {
	if a == nil {
		return b
	}
	return a
}
