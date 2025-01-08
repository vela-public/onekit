package pipekit

import (
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
)

func (s *Switch[T]) push(L *lua.LState) int {
	L.Push(lua.NewGeneric[*Switch[T]](s))
	return 1
}

func (s *Switch[T]) InvokeL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}
	for i := 1; i <= n; i++ {
		v := luakit.Check[T](L, L.Get(i))
		s.Invoke(v)
	}
	return 0
}

func (s *Switch[T]) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(s.InvokeL), true
}

func (s *Switch[T]) CaseL(L *lua.LState) int {
	cnd := cond.CheckMany(L, cond.Seek(0))
	c := s.Case(s.Cnd(cnd))
	L.Push(lua.NewGeneric[*Case[T]](c))
	return 1
}

func (s *Switch[T]) OneL(L *lua.LState) int {
	s.private.Break = true
	return s.push(L)
}
func (s *Switch[T]) DefaultL(L *lua.LState) int {
	s.private.Default = Lua[T](L, LState(L))
	return s.push(L)
}

func (s *Switch[T]) BeforeL(L *lua.LState) int {
	s.private.Before = Lua[T](L, LState(L))
	return s.push(L)
}

func (s *Switch[T]) AfterL(L *lua.LState) int {
	s.private.After = Lua[T](L, LState(L))
	return s.push(L)
}

func (s *Switch[T]) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "case":
		return lua.NewFunction(s.CaseL)
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
