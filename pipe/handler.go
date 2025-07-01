package pipe

import (
	"fmt"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
	"io"
	"runtime"
	"strconv"
)

type Handler struct {
	env    *HandleEnv
	data   any
	info   error
	invoke func(*Catalog) error
}

func (h *Handler) Data() any {
	return h.data
}

func (h *Handler) Writer(w io.Writer, c *Catalog) error {
	if c.size == 0 {
		return nil
	}

	errs := &errkit.JoinError{}
	handle := func(idx int, v []byte) {
		_, err := w.Write(v)
		c.errorf("handle[%d] write fail %v", idx, err)
	}
	for i := 0; i < c.size; i++ {
		item := c.data[i]
		if item == nil {
			continue
		}
		text, err := MarshalText(item)
		if err != nil {
			errs.Try(strconv.Itoa(i), err)
			continue
		}
		handle(i, text)
	}

	return errs.Wrap()
}

func (h *Handler) Call(fn func(v any) error, ctx *Catalog) error {
	if ctx.size == 0 {
		return nil
	}

	errs := &errkit.JoinError{}
	for i := 0; i < ctx.size; i++ {
		item := ctx.data[i]
		if item == nil {
			continue
		}
		err := fn(item)
		errs.Try(strconv.Itoa(i), err)
	}
	return errs.Wrap()
}

func (h *Handler) SafeCall(fn func(c *Catalog) error) func(ctx *Catalog) error {
	if h.env == nil || !h.env.Protect {
		return fn
	}
	return func(ctx *Catalog) error {
		defer func() {
			if e := recover(); e != nil {
				buff := make([]byte, 4096)
				n := runtime.Stack(buff, false)
				ctx.errorf("panic:%v\n%s", e, string(buff[:n]))
			}
		}()
		return fn(ctx)
	}
}

func (h *Handler) LFunc(fn *lua.LFunction, ctx *Catalog) error {
	if h.env == nil {
		return fmt.Errorf("pipe pcall env is nil")
	}

	return h.env.PCall(fn, ctx)
}

func (h *Handler) Invoke(a *Catalog) error {
	if h.invoke == nil {
		return h.info
	}
	return h.invoke(a)
}

func InvokerFunc(h *Handler, v any) {
	switch vt := v.(type) {
	case *Handler:
		h.invoke = func(a *Catalog) error {
			if vt.invoke != nil {
				return vt.invoke(a)
			}
			return fmt.Errorf("not found invoke function %v", v)
		}

	case *Chain:
		h.invoke = func(c *Catalog) error {
			return vt.InvokeGo(c).UnwrapErr()
		}
	case *Switch:
		h.invoke = func(c *Catalog) error {
			vt.Invoke(c, func(cc *Catalog) {
				cc.hijack = c.hijack
			})
			return nil
		}
	case Invoker:
		h.invoke = h.SafeCall(func(c *Catalog) error {
			return h.Call(vt.Invoke, c)
		})

	case *lua.LFunction:
		h.invoke = func(c *Catalog) error {
			return h.LFunc(vt, c)
		}

	case func(any):
		h.invoke = h.SafeCall(func(c *Catalog) error {
			return h.Call(func(v any) error { vt(v); return nil }, c)
		})

	case func(any) error:
		h.invoke = h.SafeCall(func(c *Catalog) error {
			return h.Call(func(v any) error { return vt(v) }, c)
		})

	case io.Writer:
		h.invoke = h.SafeCall(func(c *Catalog) error {
			return h.Writer(vt, c)
		})

	case lua.Invoker:
		h.invoke = func(c *Catalog) error {
			return h.Call(func(v any) error { return vt(v) }, c)
		}

	case *lua.LUserData:
		InvokerFunc(h, vt.Value)

	case lua.GenericType:
		InvokerFunc(h, vt.Unpack())

	case lua.PackType:
		InvokerFunc(h, vt.Unpack())

	default:
		h.info = fmt.Errorf("not compatible %T", v)
	}

}

func (h *Handler) prepare(v any) {
	h.data = v
	InvokerFunc(h, v)
}
