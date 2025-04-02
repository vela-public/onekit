package pipe

import (
	"fmt"
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/todo"
)

type SwitchHandler interface {
	Case(idx int, cnd *cond.Cond, v any) *Context
	NewHandler(v any, options ...func(*HandleEnv)) (r todo.Result[*Handler, error])
}

type Case struct {
	Break bool
	Cnd   *cond.Cond
	Happy SwitchHandler
	Debug SwitchHandler
}

func (c *Case) String() string                         { return fmt.Sprintf("switch.case(%s)", c.Cnd) }
func (c *Case) Type() lua.LValueType                   { return lua.LTObject }
func (c *Case) AssertFloat64() (float64, bool)         { return 0, false }
func (c *Case) AssertString() (string, bool)           { return c.Cnd.String(), true }
func (c *Case) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c *Case) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

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
	c.Debug.Case(idx, c.Cnd, v)
	return ctx, true
}

func Break(flag bool) func(c *Case) {
	return func(c *Case) { c.Break = flag }
}

func CndText(text ...string) func(c *Case) {
	return func(c *Case) { c.Cnd = cond.NewText(text...) }
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
