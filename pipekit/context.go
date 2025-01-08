package pipekit

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

type Context[T any] struct {
	size   int
	data   []T
	errs   *errkit.JoinError
	hijack struct {
		Break  bool
		Error  func(*Context[T], error)
		After  func(*Context[T])
		Before func(*Context[T])
	}
	meta Metadata
}

func (c *Context[T]) first() (t T) {
	if c.size == 0 {
		return
	}
	return c.data[0]
}

func (c *Context[T]) String() string {
	if e := c.errs.Wrap(); e != nil {
		return e.Error()
	}
	return ""
}

func (c *Context[T]) Use() []any {
	sz := len(c.data)
	if sz == 0 {
		return nil
	}
	dat := make([]any, sz)
	for i := 0; i < sz; i++ {
		dat[i] = c.data[i]
	}
	return dat
}

func (c *Context[T]) UnwrapErr() error {
	return c.errs.Wrap()
}

func (c *Context[T]) Type() lua.LValueType                   { return lua.LTObject }
func (c *Context[T]) AssertFloat64() (float64, bool)         { return 0, false }
func (c *Context[T]) AssertString() (string, bool)           { return c.String(), true }
func (c *Context[T]) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c *Context[T]) Hijack(fsm *lua.CallFrameFSM) bool      { return false }
func (c *Context[T]) Metadata() *Metadata                    { return &c.meta }
func (c *Context[T]) Index(L *lua.LState, key string) lua.LValue {
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
