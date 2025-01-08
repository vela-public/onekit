package pipekit

import (
	"fmt"
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/todo"
)

type Chain[T any] struct {
	handle []*Handler[T]
}

func (c *Chain[T]) String() string                    { return fmt.Sprintf("pipeplus.<%p>", c) }
func (c *Chain[T]) Type() lua.LValueType              { return lua.LTObject }
func (c *Chain[T]) AssertFloat64() (float64, bool)    { return float64(c.Len()), true }
func (c *Chain[T]) AssertString() (string, bool)      { return c.String(), true }
func (c *Chain[T]) Hijack(fsm *lua.CallFrameFSM) bool { return false }

func (c *Chain[T]) append(v *Handler[T]) {
	c.handle = append(c.handle, v)
}

func (c *Chain[T]) Len() int {
	return len(c.handle)
}

func (c *Chain[T]) Merge(sub *Chain[T]) {
	for _, h := range sub.handle {
		c.append(h)
	}
}

func (c *Chain[T]) Error(fn func(*Context[T], error)) func(*Context[T]) {
	return func(ctx *Context[T]) {
		ctx.hijack.Error = fn
	}
}

func (c *Chain[T]) After(fn func(*Context[T])) func(*Context[T]) {
	return func(ctx *Context[T]) {
		ctx.hijack.After = fn
	}
}

func (c *Chain[T]) Before(fn func(*Context[T])) func(*Context[T]) {
	return func(ctx *Context[T]) {
		ctx.hijack.Before = fn
	}
}

func (c *Chain[T]) Background(options ...func(*Context[T])) *Context[T] {
	ctx := &Context[T]{
		errs: errkit.Errors(),
	}

	for _, option := range options {
		option(ctx)
	}
	return ctx
}

// Do 执行管道，返回第一个错误
// example: chain.Do(chain.Background(Before,Error,After), v...)
func (c *Chain[T]) Do(ctx *Context[T], v ...T) {
	sz := len(c.handle)
	if sz == 0 {
		return
	}

	ctx.size = sz
	ctx.data = v

	if fn := ctx.hijack.Before; fn != nil {
		if fn(ctx); ctx.hijack.Break {
			return
		}
	}

	for i := 0; i < sz; i++ {
		h := c.handle[i]
		err := h.Invoke(ctx)
		if err == nil {
			continue
		}

		if fn := ctx.hijack.Error; fn != nil {
			if fn(ctx, err); ctx.hijack.Break {
				return
			}
		}

		key := fmt.Sprintf("handle.%d:%T", i, h.data)
		ctx.errs.Try(key, err)
	}

	if fn := ctx.hijack.After; fn != nil {
		fn(ctx)
	}
}

func (c *Chain[T]) Case(idx int, cnd *cond.Cond, v T) *Context[T] {
	ctx := &Context[T]{
		errs: errkit.Errors(),
		meta: Metadata{
			Switch: true,
			Cnd:    cnd,
			CaseID: idx,
		},
	}

	c.Do(ctx, v)
	return ctx
}

func (c *Chain[T]) Invoke(v ...T) *Context[T] {
	ctx := &Context[T]{
		size: len(v),
		data: v,
		errs: errkit.Errors(),
	}

	sz := len(c.handle)
	if sz == 0 {
		return ctx
	}
	for i := 0; i < sz; i++ {
		h := c.handle[i]
		if err := h.Invoke(ctx); err != nil {
			key := fmt.Sprintf("handle.%d:%T", i, h.data)
			ctx.errs.Try(key, err)
		}
	}

	return ctx
}

func (c *Chain[T]) NewHandler(v any, options ...func(*HandleEnv)) (r todo.Result[*Handler[T], error]) {
	env := NewEnv(options...)
	return c.handler(v, env)
}

func (c *Chain[T]) handler(v any, env *HandleEnv) (r todo.Result[*Handler[T], error]) {
	h, ok := v.(*Handler[T])
	if ok {
		c.append(h)
		r = todo.Ok[*Handler[T], error](h)
		return
	}

	hd := &Handler[T]{env: env}
	hd.prepare(v)
	if hd.info == nil {
		c.append(hd)
		r = todo.Ok[*Handler[T], error](hd)
		return
	}
	r = todo.Err[*Handler[T], error](hd, hd.info)
	return
}
