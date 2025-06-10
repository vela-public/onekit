package lua

import (
	"context"
	"fmt"
	"reflect"
)

type LazyLState[T any, K any] struct {
	kernel *LState
}

func NewLazyLState[T any, K any](name string, fns ...func(*Options)) *LazyLState[T, K] {
	kernel := NewStateEx(name, fns...)
	ll := &LazyLState[T, K]{
		kernel: kernel,
	}

	data := kernel.Exdata()
	if data == nil {
		return &LazyLState[T, K]{
			kernel: kernel,
		}
	}

	_, ok := data.(K)
	if ok {
		return &LazyLState[T, K]{
			kernel: kernel,
		}
	}

	ll.Err("mismatch not type of " + reflect.TypeOf((*K)(nil)).String())
	return ll
}

func Use[T any, K any](co *LState) *LazyLState[T, K] {
	return &LazyLState[T, K]{
		kernel: co,
	}
}

func (ll *LazyLState[T, K]) Err(format string, args ...any) {
	if fn := ll.kernel.Options.ErrHandle; fn != nil {
		fn(fmt.Errorf(format, args...))
	}
}

func (ll *LazyLState[T, K]) Get(idx int) LValue {
	return ll.kernel.Get(idx)
}

func (ll *LazyLState[T, K]) Push(v any) error {
	lv, ok := TypeFor(v)
	if ok {
		ll.kernel.Push(lv)
		return nil
	}
	return fmt.Errorf("not supported type %v", reflect.TypeOf(v).String())
}

func (ll *LazyLState[T, K]) Call(args, results int) {
	ll.kernel.Call(args, results)
}

func (ll *LazyLState[T, K]) NewThread() *LazyLState[T, K] {
	return Use[T, K](ll.kernel.NewThreadEx())
}

func (ll *LazyLState[T, K]) Acquire() *LState {
	return ll.kernel.Coroutine()
}

func (ll *LazyLState[T, K]) Release(co *LState) {
	ll.kernel.Keepalive(co)
}

func (ll *LazyLState[T, K]) Background() context.Context {
	ctx := ll.kernel.Context()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (ll *LazyLState[T, K]) Exit() {
	ll.kernel.terminate()
}

func (ll *LazyLState[T, K]) Terminate() {
	ll.kernel.terminate()
}

func (ll *LazyLState[T, K]) RaiseError(format string, args ...any) {
	ll.kernel.RaiseError(format, args...)
}

func (ll *LazyLState[T, K]) Unwrap() *LState {
	return ll.kernel
}

func (ll *LazyLState[T, K]) PanicErr(e error) {
	if e != nil {
		ll.kernel.RaiseError("%v", e)
	}
}

func (ll *LazyLState[T, k]) CallByParam(cp P, args ...LValue) error {
	return ll.kernel.CallByParam(cp, args...)
}

func (ll *LazyLState[T, k]) DoString(text string) error {
	return ll.kernel.DoString(text)
}

func (ll *LazyLState[T, k]) DoFile(filename string) error {
	return ll.kernel.DoFile(filename)
}

func (ll *LazyLState[T, K]) Exdata() K {
	return ll.kernel.Exdata().(K)
}

func (ll *LazyLState[T, K]) Exdata2() T {
	return ll.kernel.Exdata2().(T)
}

func (ll *LazyLState[T, K]) SetExdata2(data T) T {
	return ll.kernel.Exdata2().(T)
}
