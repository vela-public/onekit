package tern

func T[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

func F[T any](fn func() error, a, b T) T {
	if err := fn(); err != nil {
		return b
	}
	return a
}
