package pipe

import (
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/lua"
)

func (s *Switch) push(L *lua.LState) int {
	L.Push(s)
	return 1
}

func (s *Switch) InvokeL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}
	for i := 1; i <= n; i++ {
		v := L.Get(i)
		s.Invoke(v)
	}
	return 0
}

func (s *Switch) CaseTextL(L *lua.LState) int {
	cnd := cond.CheckMany(L, cond.Seek(0))
	c := s.Case(Cnd(cnd))
	L.Push(c)
	return 1
}

func (s *Switch) CaseCEL(L *lua.LState) int {
	cnd := cond.CheckMany(L, cond.Seek(0))
	c := s.Case(Cnd(cnd))
	L.Push(c)
	return 1
}

func (s *Switch) OneL(L *lua.LState) int {
	s.Break = true
	return s.push(L)
}

func (s *Switch) DefaultL(L *lua.LState) int {
	s.Default = Lua(L, LState(L))
	return s.push(L)
}

func (s *Switch) BeforeL(L *lua.LState) int {
	s.Before = Lua(L, LState(L), Seek(1))
	return s.push(L)
}

func (s *Switch) AfterL(L *lua.LState) int {
	s.After = Lua(L, LState(L))
	return s.push(L)
}

func (s *Switch) Index(L *lua.LState, key string) lua.LValue {
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
