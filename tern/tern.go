package tern

func IF[T any](condition bool, a, b T) T {
	if condition {
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
