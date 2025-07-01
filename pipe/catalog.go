package pipe

import (
	"fmt"
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
	Break     bool
	Error     func(error)
	Exception func(*Catalog, error)
	After     func(*Catalog)
	Before    func(*Catalog)
}

type Catalog struct {
	size   int
	unary  any
	data   []any
	errs   *errkit.JoinError
	hijack Hijack
	meta   Metadata
}

func (c *Catalog) String() string {
	if c.errs.Len() > 0 {
		return c.errs.Error()
	}
	return ""
}

func (c *Catalog) err(e error) {
	if c.hijack.Error != nil {
		c.hijack.Error(e)
	}

	if c.hijack.Exception != nil {
		c.hijack.Exception(c, e)
	}
}

func (c *Catalog) errorf(format string, v ...any) {
	e := fmt.Errorf(format, v...)
	c.err(e)
}

func (c *Catalog) UnwrapErr() error {
	return c.errs.Wrap()
}

func (c *Catalog) Type() lua.LValueType                   { return lua.LTObject }
func (c *Catalog) AssertFloat64() (float64, bool)         { return 0, false }
func (c *Catalog) AssertString() (string, bool)           { return c.String(), true }
func (c *Catalog) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c *Catalog) Hijack(fsm *lua.CallFrameFSM) bool      { return false }
func (c *Catalog) Metadata() *Metadata                    { return &c.meta }
func (c *Catalog) Index(L *lua.LState, key string) lua.LValue {
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

func Error(fn func(error)) func(*Catalog) {
	return func(ctx *Catalog) {
		ctx.hijack.Error = fn
	}
}

func Param(v ...any) func(ctx *Catalog) {
	return func(ctx *Catalog) {
		ctx.size = len(v)
		ctx.data = v
	}
}

func Exception(fn func(*Catalog, error)) func(*Catalog) {
	return func(ctx *Catalog) {
		ctx.hijack.Exception = fn
	}
}

func After(fn func(*Catalog)) func(*Catalog) {
	return func(ctx *Catalog) {
		ctx.hijack.After = fn
	}
}

func Before(fn func(*Catalog)) func(*Catalog) {
	return func(ctx *Catalog) {
		ctx.hijack.Before = fn
	}
}

func meta(v Metadata) func(*Catalog) {
	return func(ctx *Catalog) {
		ctx.meta = v
	}
}

func Background(options ...func(*Catalog)) *Catalog {
	c := &Catalog{
		errs: errkit.Errors(),
	}

	for _, option := range options {
		option(c)
	}
	return c
}

func NewCatalog(v ...any) func(...func(*Catalog)) *Catalog {
	sz := len(v)
	ctx := &Catalog{
		size: sz,
		data: v,
		errs: errkit.Errors(),
	}

	return func(options ...func(*Catalog)) *Catalog {
		for _, option := range options {
			option(ctx)
		}
		return ctx
	}
}
