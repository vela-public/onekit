package pipe

import (
	"fmt"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/todo"
)

type Chain struct {
	handle []*Handler
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

func (c *Chain) TryCatch(ca *Catalog, idx int, err error) (stop bool) {
	if err == nil {
		return
	}

	if ca.hijack.Break {
		stop = true
	}

	key := fmt.Sprintf("handle.%d:%T", idx, c.handle[idx].data)
	ca.errs.Try(key, err)

	if ca.meta.Switch {
		ca.errorf("switch[%d][%s] invoke fail %v", ca.meta.CaseID, ca.meta.Cnd, err)
	} else {
		ca.errorf("handle[%d] invoke fail %v", idx, err)
	}

	exception := ca.hijack.Exception
	if exception != nil {
		exception(ca, err)
	}

	return
}

func (c *Chain) Execute(ctx *Catalog) {
	sz := len(c.handle)
	if sz == 0 {
		return
	}

	if fn := ctx.hijack.Before; fn != nil {
		if fn(ctx); ctx.hijack.Break {
			return
		}
	}

	for i := 0; i < sz; i++ {
		h := c.handle[i]
		err := h.Invoke(ctx)
		c.TryCatch(ctx, i, err)
	}

	if fn := ctx.hijack.After; fn != nil {
		fn(ctx)
	}
}

func (c *Chain) Do(options ...func(*Catalog)) *Catalog {
	ctx := &Catalog{}
	sz := len(options)
	if sz == 0 {
		return ctx
	}

	for i := 0; i < sz; i++ {
		option := options[i]
		option(ctx)
	}
	ctx.errs = errkit.Errors()
	c.Execute(ctx)
	return ctx
}

func (c *Chain) Invokes(v []any, more ...func(ctx *Catalog)) {
	sz := len(c.handle)
	if sz == 0 {
		return
	}

	ctx := NewCatalog(v...)(more...)
	c.Execute(ctx)
}

func (c *Chain) InvokeGo(v ...any) *Catalog {
	ca := NewCatalog(v...)()

	sz := len(c.handle)
	if sz == 0 {
		return ca
	}

	c.Execute(ca)
	return ca
}

func (c *Chain) Invoke(v any, more ...func(ctx *Catalog)) {
	sz := len(c.handle)
	if sz == 0 {
		return
	}

	ca := NewCatalog(v)(more...)
	c.Execute(ca)
}

func (c *Chain) NewHandler(v any, options ...func(*HandleEnv)) (r todo.Result[*Handler, error]) {
	env := NewEnv(options...)
	return c.handler(v, env)
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
