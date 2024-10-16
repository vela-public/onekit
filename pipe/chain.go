package pipe

import "C"
import (
	"fmt"
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/todo"
)

type Chain struct {
	handle []*Handler
}

func (c *Chain) append(v *Handler) {
	c.handle = append(c.handle, v)
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

func (c *Chain) NewHandler(v any, options ...func(*HandleEnv)) (r todo.Result[*Handler, error]) {
	env := NewEnv(options...)
	return c.handler(v, env)
}

func (c *Chain) handler(v any, env *HandleEnv) (r todo.Result[*Handler, error]) {
	h, ok := v.(*Handler)
	if ok {
		c.append(h)
		r.Value = h
		r.Ok = true
	}

	hd := &Handler{env: env}
	hd.prepare(v)
	if hd.info == nil {
		c.append(hd)
		r.Value = hd
		r.Ok = true
		return
	}
	r.Value = hd
	r.Error = hd.info
	r.Ok = false
	return
}
