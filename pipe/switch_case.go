package pipe

import (
	"github.com/vela-public/onekit/cond"
)

type Case struct {
	Switch *Switch
	Break  bool
	Cnd    *cond.Cond
	Happy  *Chain
	Debug  *Chain
}

func (c *Case) Field(key string) string {
	switch key {
	case "raw":
		return c.Cnd.String()
	}
	return ""
}

func (c *Case) Match(idx int, v any) (*Context, bool) {
	if c.Cnd == nil {
		ctx := &Context{
			meta: Metadata{
				Switch: true,
				CaseID: idx,
			},
		}
		return ctx, false
	}

	if !c.Cnd.Match(v) {
		ctx := &Context{
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

func Break(flag bool) func(c *Case) {
	return func(c *Case) { c.Break = flag }
}

func CndText(text ...string) func(c *Case) {
	return func(c *Case) { c.Cnd = cond.New(text...) }
}
func Cnd(cnd *cond.Cond) func(c *Case) {
	return func(c *Case) { c.Cnd = cnd }
}

func HappyChain(h *Chain) func(*Case) {
	return func(c *Case) { c.Happy = h }
}

func DebugChain(h *Chain) func(*Case) {
	return func(c *Case) { c.Debug = h }
}

func Happy(v any, options ...func(*HandleEnv)) func(c *Case) {
	return func(c *Case) {
		c.Happy.NewHandler(v, options...)
	}
}

func Debug(v any, options ...func(*HandleEnv)) func(c *Case) {
	return func(c *Case) { c.Debug.NewHandler(v, options...) }
}
