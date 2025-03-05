package pipekit

import (
	"fmt"
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/lua"
)

type Case[T any] struct {
	Switch *Switch[T]
	Break  bool
	Cnd    *cond.Cond
	Happy  *Chain[T]
	Debug  *Chain[*Context[T]]
}

func (c *Case[T]) String() string                         { return fmt.Sprintf("case(%s)", c.Cnd) }
func (c *Case[T]) Type() lua.LValueType                   { return lua.LTObject }
func (c *Case[T]) AssertFloat64() (float64, bool)         { return 0, false }
func (c *Case[T]) AssertString() (string, bool)           { return c.Cnd.String(), true }
func (c *Case[T]) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c *Case[T]) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

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
