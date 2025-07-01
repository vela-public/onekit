package pipe

import "fmt"

type HandleType uint8

type Invoker interface {
	Invoke(v interface{}) error
}

type InvokerT[T any] interface {
	Invoke(v T) error
}

type Bytes interface {
	Bytes() []byte
}

func Invoke[T any](fn func(T)) func(any) {
	return func(v any) {
		t, ok := v.(T)
		if ok {
			fn(t)
		}
	}
}

func InvokeE[T any](fn func(T) error) func(any) error {
	return func(v any) error {
		var t T
		var ok bool
		if t, ok = v.(T); ok {
			return fn(t)
		}
		return fmt.Errorf("bad type %T must %T", v, t)
	}
}
