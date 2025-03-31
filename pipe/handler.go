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
	invoke func(*Context) error
}

func (h *Handler) Data() any {
	return h.data
}

func (h *Handler) Writer(w io.Writer, c *Context) error {
	if c.size == 0 {
		return nil
	}

	errs := &errkit.JoinError{}
	handle := func(idx int, v []byte) {
		_, err := w.Write(v)
		if err != nil {
			errs.Try(strconv.Itoa(idx), err)
		}
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

func (h *Handler) SafeCall(fn func(v any) error, ctx *Context) error {
	if ctx.size == 0 {
		return nil
	}

	errs := &errkit.JoinError{}
	for i := 0; i < ctx.size; i++ {
		item := ctx.data[i]
		if item == nil {
			continue
		}
		errs.Try(strconv.Itoa(i), fn(item))
	}
	return errs.Wrap()
}

func (h *Handler) Protect() {
	if !h.env.Protect {
		return
	}

	r := recover()
	if r != nil {
		buff := make([]byte, 4096)
		n := runtime.Stack(buff, false)
		fmt.Println(string(buff[:n]))
	}
}

func (h *Handler) LFunc(fn *lua.LFunction, ctx *Context) error {
	defer h.Protect()

	if h.env == nil {
		return fmt.Errorf("pipe pcall env is nil")
	}

	return h.env.PCall(fn, ctx)
}

func (h *Handler) Invoke(a *Context) error {
	if h.invoke == nil {
		return h.info
	}
	return h.invoke(a)
}

func InvokerFunc(h *Handler, v any) {
	switch elem := v.(type) {
	case *Chain:
		h.invoke = func(a *Context) error {
			elem.Do(a, a.data...)
			return nil
		}
	case *Switch:
		h.invoke = func(a *Context) error {
			elem.Invoke(a)
			return nil
		}

	case io.Writer:
		h.invoke = func(a *Context) error {
			return h.Writer(elem, a)
		}
	case Invoker:
		h.invoke = func(a *Context) error {
			return h.SafeCall(elem.Invoke, a)
		}

	case *lua.LFunction:
		h.invoke = func(a *Context) error {
			return h.LFunc(elem, a)
		}

	case lua.GoFuncErr:
		h.invoke = func(a *Context) error {
			defer h.Protect()
			return elem(a.data...)
		}

	case lua.GoFuncStr:
		h.invoke = func(a *Context) error {
			defer h.Protect()
			_ = elem(a.data...)
			return nil
		}
	case lua.GoFuncInt:
		h.invoke = func(a *Context) error {
			defer h.Protect()
			_ = elem(a.data...)
			return nil
		}
	case func(any):
		h.invoke = func(a *Context) error {
			return h.SafeCall(func(v any) error { elem(v); return nil }, a)
		}
	case func(any) error:
		h.invoke = func(a *Context) error {
			return h.SafeCall(func(v any) error { return elem(v) }, a)
		}
	case *lua.LUserData:
		InvokerFunc(h, elem.Value)

	case lua.GenericType:
		InvokerFunc(h, elem.Unpack())
	case lua.Invoker:
		h.invoke = func(a *Context) error {
			return h.SafeCall(func(v any) error { return elem(v) }, a)
		}
	case lua.PackType:
		InvokerFunc(h, elem.Unpack())
	default:
		h.info = fmt.Errorf("not compatible %T", v)
	}

}

func (h *Handler) prepare(v any) {
	h.data = v
	InvokerFunc(h, v)
}
