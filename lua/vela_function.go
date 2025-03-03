package lua

import "fmt"

type GoFunction[T any] func(T)

func (fn GoFunction[T]) String() string                     { return fmt.Sprintf("Go function error: %p", fn) }
func (fn GoFunction[T]) Type() LValueType                   { return LTGoFunction }
func (fn GoFunction[T]) AssertFloat64() (float64, bool)     { return 0, false }
func (fn GoFunction[T]) AssertString() (string, bool)       { return "", false }
func (fn GoFunction[T]) AssertFunction() (*LFunction, bool) { return nil, false }
func (fn GoFunction[T]) Hijack(*CallFrameFSM) bool          { return false }

type GoFuncErr func(...interface{}) error

func (ge GoFuncErr) String() string                     { return fmt.Sprintf("Go function error: %p", ge) }
func (ge GoFuncErr) Type() LValueType                   { return LTGoFuncErr }
func (ge GoFuncErr) AssertFloat64() (float64, bool)     { return 0, false }
func (ge GoFuncErr) AssertString() (string, bool)       { return "", false }
func (ge GoFuncErr) AssertFunction() (*LFunction, bool) { return nil, false }
func (ge GoFuncErr) Hijack(*CallFrameFSM) bool          { return false }

type GoFuncStr func(...interface{}) string

func (gs GoFuncStr) String() string                     { return fmt.Sprintf("Go fucntion string: %p", gs) }
func (gs GoFuncStr) Type() LValueType                   { return LTGoFuncStr }
func (gs GoFuncStr) AssertFloat64() (float64, bool)     { return 0, false }
func (gs GoFuncStr) AssertString() (string, bool)       { return "", false }
func (gs GoFuncStr) AssertFunction() (*LFunction, bool) { return nil, false }
func (gs GoFuncStr) Hijack(*CallFrameFSM) bool          { return false }

type GoFuncInt func(...interface{}) int

func (gi GoFuncInt) String() string                     { return fmt.Sprintf("Go function int : %p", gi) }
func (gi GoFuncInt) Type() LValueType                   { return LTGoFuncInt }
func (gi GoFuncInt) AssertFloat64() (float64, bool)     { return 0, false }
func (gi GoFuncInt) AssertString() (string, bool)       { return "", false }
func (gi GoFuncInt) AssertFunction() (*LFunction, bool) { return nil, false }
func (gi GoFuncInt) Hijack(*CallFrameFSM) bool          { return false }

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
