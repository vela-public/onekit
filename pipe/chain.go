package pipe

import (
	"fmt"
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/todo"
)

type Chain struct {
	handle []*Handler

	private struct {
		ErrHandle []*Handler
	}
}

func (c *Chain) append(v *Handler) {
	c.handle = append(c.handle, v)
}

func (c *Chain) Len() int {
	return len(c.handle)
}

func (c *Chain) Merge(sub *Chain) {
	for _, h := range sub.handle {
		c.append(h)
	}
}

func (c *Chain) error(err error) {
	if err == nil {
		return
	}

	sz := len(c.private.ErrHandle)
	if sz == 0 {
		return
	}

	ctx := &Context{
		size: 1,
		data: []any{any(err)},
	}

	for i := 0; i < sz; i++ {
		h := c.private.ErrHandle[i]
		_ = h.invoke(ctx)
	}
}

func (c *Chain) Do(ctx *Context, v ...any) {
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

func (c *Chain) Case(idx int, cnd *cond.Cond, v any) *Context {
	ctx := &Context{
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

func (c *Chain) Invoke(v ...any) *Context {
	ctx := &Context{
		size: len(v),
		data: v,
		errs: errkit.Errors(),
	}
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("pipe invoke panic %v\n%s", e, libkit.StackTrace[string](4096, true))
		}
	}()

	sz := len(c.handle)
	if sz == 0 {
		return ctx
	}
	for i := 0; i < sz; i++ {
		h := c.handle[i]
		if err := h.Invoke(ctx); err != nil {
			key := fmt.Sprintf("handle.%d:%T", i, h.data)
			ctx.errs.Try(key, err)
			c.error(err)
		}
	}
	return ctx
}

func (c *Chain) NewHandler(v any, options ...func(*HandleEnv)) (r todo.Result[*Handler, error]) {
	env := NewEnv(options...)
	return c.handler(v, env)
}

func (c *Chain) NewErrorHandler(v any, options ...func(*HandleEnv)) error {
	env := NewEnv(options...)
	h, ok := v.(*Handler)
	if ok {
		c.private.ErrHandle = append(c.private.ErrHandle, h)
		return nil
	}

	hd := &Handler{env: env}
	hd.prepare(v)
	if hd.info == nil {
		c.private.ErrHandle = append(c.private.ErrHandle, h)
		return nil
	}
	return hd.info
}

func (c *Chain) handler(v any, env *HandleEnv) (r todo.Result[*Handler, error]) {
	h, ok := v.(*Handler)
	if ok {
		c.append(h)
		r = todo.Ok[*Handler, error](h)
		return
	}

	hd := &Handler{env: env}
	hd.prepare(v)
	if hd.info == nil {
		c.append(hd)
		r = todo.Ok[*Handler, error](hd)
		return
	}
	r = todo.Err[*Handler, error](hd, hd.info)
	return
}
