package pipe

import (
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
)

type Metadata struct {
	Switch bool
	Cnd    *cond.Cond
	CaseID int
	Return libkit.DataKV[string, any]
}

type Hijack struct {
	Break  bool
	Error  func(*Context, error)
	After  func(*Context)
	Before func(*Context)
}

type Context struct {
	size   int
	data   []any
	errs   *errkit.JoinError
	hijack Hijack
	meta   Metadata
}

func (c *Context) String() string {
	if e := c.errs.Wrap(); e != nil {
		return e.Error()
	}
	return ""
}

func (c *Context) Type() lua.LValueType                   { return lua.LTObject }
func (c *Context) AssertFloat64() (float64, bool)         { return 0, false }
func (c *Context) AssertString() (string, bool)           { return c.String(), true }
func (c *Context) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c *Context) Hijack(fsm *lua.CallFrameFSM) bool      { return false }
func (c *Context) Metadata() *Metadata                    { return &c.meta }
func (c *Context) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "size":
		return lua.LInt(c.size)
	case "err":
		if e := c.errs.Wrap(); e != nil {
			return lua.LString(e.Error())
		}
		return lua.LNil
	case "case_id":
		return lua.LInt(c.meta.CaseID)
	case "case_cnd":
		if c.meta.Switch && c.meta.Cnd != nil {
			return lua.LString(c.meta.Cnd.String())
		}
		return lua.LNil
	}

	return lua.LNil
}

func Error(fn func(*Context, error)) func(*Context) {
	return func(ctx *Context) {
		ctx.hijack.Error = fn
	}
}

func After(fn func(*Context)) func(*Context) {
	return func(ctx *Context) {
		ctx.hijack.After = fn
	}
}

func Before(fn func(*Context)) func(*Context) {
	return func(ctx *Context) {
		ctx.hijack.Before = fn
	}
}

func Background(options ...func(*Context)) *Context {
	c := &Context{
		errs: errkit.Errors(),
	}

	for _, option := range options {
		option(c)
	}
	return c
}
