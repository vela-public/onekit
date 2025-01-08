package pipekit

import (
	"github.com/vela-public/onekit/cond"
)

type Case[T any] struct {
	Switch *Switch[T]
	Break  bool
	Cnd    *cond.Cond
	Happy  *Chain[T]
	Debug  *Chain[*Context[T]]
}

func (c *Case[T]) Field(key string) string {
	switch key {
	case "raw":
		return c.Cnd.String()
	}
	return ""
}

func (c *Case[T]) Match(idx int, v T) (*Context[T], bool) {
	if c.Cnd == nil {
		ctx := &Context[T]{
			meta: Metadata{
				Switch: true,
				CaseID: idx,
			},
		}
		return ctx, false
	}

	if !c.Cnd.Match(v) {
		ctx := &Context[T]{
			meta: Metadata{
				Switch: true,
				CaseID: idx,
				Cnd:    c.Cnd,
			},
		}
		return ctx, false
	}

	ctx := c.Happy.Case(idx, c.Cnd, v)
	ctx.meta.Cnd = c.Cnd
	ctx.meta.CaseID = idx
	ctx.meta.Switch = true
	c.Debug.Invoke(ctx)
	return ctx, true
}
