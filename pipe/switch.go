package pipe

import (
	"github.com/vela-public/onekit/lua"
)

type Switch struct {
	Debug   bool
	Break   bool
	Cases   []*Case
	Default *Chain
	Before  *Chain
	After   *Chain
}

func (s *Switch) String() string                         { return "switch" }
func (s *Switch) Type() lua.LValueType                   { return lua.LTObject }
func (s *Switch) AssertFloat64() (float64, bool)         { return float64(len(s.Cases)), true }
func (s *Switch) AssertString() (string, bool)           { return s.String(), true }
func (s *Switch) AssertFunction() (*lua.LFunction, bool) { return lua.NewFunction(s.InvokeL), true }
func (s *Switch) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func More(ctx *Context, more ...func(*Context)) {
	sz := len(more)
	if sz == 0 {
		return
	}

	for i := 0; i < sz; i++ {
		more[i](ctx)
	}
}

func (s *Switch) OnBefore(v any, options ...func(*HandleEnv)) {
	s.Before.NewHandler(v, options...)
}

func (s *Switch) OnAfter(v any, options ...func(*HandleEnv)) {
	s.After.NewHandler(v, options...)
}

func (s *Switch) NotFound(v any, options ...func(*HandleEnv)) {
	s.Default.NewHandler(v, options...)
}

func (s *Switch) Invoke(v any, more ...func(*Context)) {
	s.Before.Invoke(v)

	sz := len(s.Cases)
	if sz == 0 {
		return
	}

	hit := false
	for i := 0; i < sz; i++ {
		c := s.Cases[i]
		ctx, ok := c.Match(i, v)
		if !ok {
			continue
		}

		hit = true
		More(ctx, more...)
		if s.Break || c.Break {
			break
		}

	}

	if !hit {
		s.Default.Invoke(v)
	}
	s.After.Invoke(v)
}

func (s *Switch) Case(options ...func(*Case)) *Case {
	c := &Case{
		Switch: s,
		Happy:  NewChain(),
		Debug:  NewChain(),
		Break:  true,
	}
	for _, opt := range options {
		opt(c)
	}
	s.Cases = append(s.Cases, c)
	return c
}

func NewSwitch() *Switch {
	return &Switch{
		Default: NewChain(),
		Before:  NewChain(),
		After:   NewChain(),
	}
}

/*

    s := Switch().one()
	s.case("name == google").invoke(func(x) end).more().break()

	s := NewSwitch()
	s.case("name == google").invoke().one().debug()
	s.case("name == 122333").pipe()
	s.case("name == application").pipe()
	s.case("name == session").pipe()

*/
