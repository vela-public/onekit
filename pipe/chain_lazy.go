package pipe

import (
	"fmt"
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/todo"
)

type LazyChain[T any] struct {
	Chain *Chain
}

func (zc *LazyChain[T]) String() string                         { return "lazy.chain" }
func (zc *LazyChain[T]) Type() lua.LValueType                   { return lua.LTObject }
func (zc *LazyChain[T]) AssertFloat64() (float64, bool)         { return float64(zc.Chain.Len()), true }
func (zc *LazyChain[T]) AssertString() (string, bool)           { return "", false }
func (zc *LazyChain[T]) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (zc *LazyChain[T]) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (zc *LazyChain[T]) NewHandler(v any, option ...func(env *HandleEnv)) (r todo.Result[*Handler, error]) {
	switch dat := v.(type) {
	case func(T) error:
		fn := func(v any) error {
			if data, ok := v.(T); ok {
				return dat(data)
			}
			return fmt.Errorf("data type is not %T", *new(T))
		}
		return zc.Chain.NewHandler(fn, option...)

	case func(T):
		return zc.Chain.NewHandler(dat, option...)

	case InvokerT[T]:
		fn := func(v any) error {
			if data, ok := v.(T); ok {
				return dat.Invoke(data)
			}
			return fmt.Errorf("data type is not %T", *new(T))
		}
		return zc.Chain.NewHandler(fn, option...)

	default:
		return zc.Chain.NewHandler(v, option...)
	}
}

func (zc *LazyChain[T]) Merge(v any) {
	switch sub := v.(type) {
	case *LazyChain[T]:
		zc.Chain.Merge(sub.Chain)
	case *Chain:
		zc.Chain.Merge(sub)
	}
}

func (zc *LazyChain[T]) Do(ctx *Context, v any) {
	zc.Chain.Do(ctx, v)
}

func (zc *LazyChain[T]) Invoke(v ...any) {
	zc.Chain.Invoke(v...)
}

func (zc *LazyChain[T]) Case(idx int, cnd *cond.Cond, v any) *Context {
	ctx := &Context{
		errs: errkit.Errors(),
		meta: Metadata{
			Switch: true,
			Cnd:    cnd,
			CaseID: idx,
		},
	}

	zc.Chain.Do(ctx, v)
	return ctx
}

func NewLazyChain[T any]() *LazyChain[T] {
	return &LazyChain[T]{
		Chain: NewChain(),
	}
}

func LuaLazyChain[T any](L *lua.LState, options ...func(*HandleEnv)) *LazyChain[T] {
	env := NewEnv(options...)

	zc := NewLazyChain[T]()

	if env.Seek == 0 {
		env.Seek = 1
	}

	n := L.GetTop()
	if n-env.Seek < 0 {
		return zc
	}

	for idx := env.Seek; idx <= n; idx++ {
		zc.NewHandler(L.Get(idx), Clone(env))
	}
	return zc
}
