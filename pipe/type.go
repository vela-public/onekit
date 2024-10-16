package pipe

import "fmt"

const (
	Single HandleType = iota + 1
	ReuseCo
)

type HandleType uint8

type Invoker interface {
	Invoke(v interface{}) error
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
