package lua

import "fmt"

type GoFunction[T any] func(T)

func (fn GoFunction[T]) String() string                     { return fmt.Sprintf("Go function error: %p", fn) }
func (fn GoFunction[T]) Type() LValueType                   { return LTGoFunction }
func (fn GoFunction[T]) AssertFloat64() (float64, bool)     { return 0, false }
func (fn GoFunction[T]) AssertString() (string, bool)       { return "", false }
func (fn GoFunction[T]) AssertFunction() (*LFunction, bool) { return nil, false }
func (fn GoFunction[T]) Hijack(*CallFrameFSM) bool          { return false }

type Invoker func(any) error

func (i Invoker) String() string                     { return fmt.Sprintf("invoker function int : %p", i) }
func (i Invoker) Type() LValueType                   { return LTInvoker }
func (i Invoker) AssertFloat64() (float64, bool)     { return 0, false }
func (i Invoker) AssertString() (string, bool)       { return "", false }
func (i Invoker) AssertFunction() (*LFunction, bool) { return nil, false }
func (i Invoker) Hijack(*CallFrameFSM) bool          { return false }

type InvokerOf[T any] func(T) error

func (i InvokerOf[T]) String() string                     { return fmt.Sprintf("invoker function int : %p", i) }
func (i InvokerOf[T]) Type() LValueType                   { return LTInvoker }
func (i InvokerOf[T]) AssertFloat64() (float64, bool)     { return 0, false }
func (i InvokerOf[T]) AssertString() (string, bool)       { return "", false }
func (i InvokerOf[T]) AssertFunction() (*LFunction, bool) { return nil, false }
func (i InvokerOf[T]) Hijack(*CallFrameFSM) bool          { return false }

type GoCond[T any] func(T) bool

func (gc GoCond[T]) String() string                     { return fmt.Sprintf("function: %T", gc) }
func (gc GoCond[T]) Type() LValueType                   { return LTGoCond }
func (gc GoCond[T]) AssertFloat64() (float64, bool)     { return 0, false }
func (gc GoCond[T]) AssertString() (string, bool)       { return "", false }
func (gc GoCond[T]) AssertFunction() (*LFunction, bool) { return nil, false }
func (gc GoCond[T]) Hijack(*CallFrameFSM) bool          { return false }

func LazyInvoker[T any](fn func(T)) Invoker {
	return func(v any) error {
		t, ok := v.(T)
		if !ok {
			return fmt.Errorf("invalid type: %T", v)
		}
		fn(t)
		return nil
	}
}

func LazyInvokerE[T any](fn func(T) error) Invoker {
	return func(v any) error {
		t, ok := v.(T)
		if !ok {
			return fmt.Errorf("invalid type: %T", v)
		}
		return fn(t)
	}
}
