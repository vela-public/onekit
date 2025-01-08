package pipekit

import "github.com/vela-public/onekit/cond"

type Switch[T any] struct {
	private struct {
		Debug   bool
		Break   bool
		Default *Chain[T]
		Before  *Chain[T]
		After   *Chain[T]
	}

	Cases []*Case[T]
}

func (s *Switch[T]) More(ctx *Context[T], more ...func(*Context[T])) {
	sz := len(more)
	if sz == 0 {
		return
	}

	for i := 0; i < sz; i++ {
		more[i](ctx)
	}
}

func (s *Switch[T]) OnBefore(v any, options ...func(*HandleEnv)) {
	s.private.Before.NewHandler(v, options...)
}

func (s *Switch[T]) OnAfter(v any, options ...func(*HandleEnv)) {
	s.private.After.NewHandler(v, options...)
}

func (s *Switch[T]) NotFound(v any, options ...func(*HandleEnv)) {
	s.private.Default.NewHandler(v, options...)
}

func (s *Switch[T]) Invoke(v T, more ...func(*Context[T])) {
	s.private.Before.Invoke(v)

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
		s.More(ctx, more...)
		if s.private.Break || c.Break {
			break
		}

	}

	if !hit {
		s.private.Default.Invoke(v)
	}
	s.private.After.Invoke(v)
}

func (s *Switch[T]) Break(flag bool) func(c *Case[T]) {
	return func(c *Case[T]) { c.Break = flag }
}

func (s *Switch[T]) CndText(text ...string) func(c *Case[T]) {
	return func(c *Case[T]) { c.Cnd = cond.New(text...) }
}
func (s *Switch[T]) Cnd(cnd *cond.Cond) func(c *Case[T]) {
	return func(c *Case[T]) { c.Cnd = cnd }
}

func (s *Switch[T]) HappyChain(h *Chain[T]) func(*Case[T]) {
	return func(c *Case[T]) { c.Happy = h }
}

func (s *Switch[T]) DebugChain(h *Chain[*Context[T]]) func(*Case[T]) {
	return func(c *Case[T]) { c.Debug = h }
}

func (s *Switch[T]) Happy(v func(T), options ...func(*HandleEnv)) func(c *Case[T]) {
	return func(c *Case[T]) {
		c.Happy.NewHandler(v, options...)
	}
}

func (s *Switch[T]) Default(v func(T), options ...func(*HandleEnv)) {
	s.private.Default.NewHandler(v, options...)
}

func (s *Switch[T]) Debug(v func(ctx *Context[T]), options ...func(*HandleEnv)) func(c *Case[T]) {
	return func(c *Case[T]) { c.Debug.NewHandler(v, options...) }
}

func (s *Switch[T]) Case(options ...func(*Case[T])) *Case[T] {
	c := &Case[T]{
		Switch: s,
		Happy:  NewChain[T](),
		Debug:  NewChain[*Context[T]](),
		Break:  true,
	}
	for _, opt := range options {
		opt(c)
	}
	s.Cases = append(s.Cases, c)
	return c
}

func NewSwitch[T any]() *Switch[T] {
	s := &Switch[T]{}
	s.private.Default = NewChain[T]()
	s.private.Before = NewChain[T]()
	s.private.After = NewChain[T]()
	s.private.Debug = false
	s.private.Break = false
	return s
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
