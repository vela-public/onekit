package pipe

import (
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/lua"
)

type LazySwitch[T any] struct {
	ref *Switch
}

func (s *LazySwitch[T]) String() string                    { return "switch" }
func (s *LazySwitch[T]) Type() lua.LValueType              { return lua.LTObject }
func (s *LazySwitch[T]) AssertFloat64() (float64, bool)    { return float64(len(s.ref.Cases)), true }
func (s *LazySwitch[T]) AssertString() (string, bool)      { return s.String(), true }
func (s *LazySwitch[T]) Hijack(fsm *lua.CallFrameFSM) bool { return false }
func (s *LazySwitch[T]) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(s.ref.InvokeL), true
}

func (s *LazySwitch[T]) Invoke(v T, more ...func(*Catalog)) {
	s.ref.Invoke(v, more...)
}

func (s *LazySwitch[T]) Case(options ...func(*Case)) *Case {
	c := &Case{
		Happy: NewLazyChain[T](),
		Debug: NewLazyChain[T](),
		Break: true,
	}
	for _, opt := range options {
		opt(c)
	}
	s.ref.Cases = append(s.ref.Cases, c)
	return c
}

func (s *LazySwitch[T]) push(L *lua.LState) int {
	L.Push(s)
	return 1
}

func (s *LazySwitch[T]) InvokeL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}
	for i := 1; i <= n; i++ {
		v := L.Get(i)
		s.ref.Invoke(v)
	}
	return 0
}

func (s *LazySwitch[T]) CaseTextL(L *lua.LState) int {
	cnd := cond.CheckMany(L, cond.Seek(0))
	c := s.Case(Cnd(cnd))
	L.Push(c)
	return 1
}

func (s *LazySwitch[T]) CaseCEL(L *lua.LState) int {
	cnd := cond.CheckMany(L, cond.Seek(0))
	c := s.Case(Cnd(cnd))
	L.Push(c)
	return 1
}

func (s *LazySwitch[T]) OneL(L *lua.LState) int {
	s.ref.Break = true
	return s.push(L)
}

func (s *LazySwitch[T]) DefaultL(L *lua.LState) int {
	s.ref.Default = Lua(L, LState(L))
	return s.push(L)
}

func (s *LazySwitch[T]) BeforeL(L *lua.LState) int {
	s.ref.Before = Lua(L, LState(L), Seek(1))
	return s.push(L)
}

func (s *LazySwitch[T]) AfterL(L *lua.LState) int {
	s.ref.After = Lua(L, LState(L))
	return s.push(L)
}

func (s *LazySwitch[T]) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "case":
		return lua.NewFunction(s.CaseTextL)
	case "case_cel":
		return lua.NewFunction(s.CaseCEL)
	case "one":
		return lua.NewFunction(s.OneL)
	case "before":
		return lua.NewFunction(s.BeforeL)
	case "after":
		return lua.NewFunction(s.AfterL)
	case "default":
		return lua.NewFunction(s.DefaultL)
	default:
		return nil
	}

}

func NewLazySwitch[T any]() *LazySwitch[T] {
	return &LazySwitch[T]{
		ref: &Switch{
			Default: NewLazyChain[T](),
			Before:  NewChain(),
			After:   NewChain(),
		},
	}
}
