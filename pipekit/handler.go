package pipekit

import (
	"fmt"
	"github.com/vela-public/onekit/deflect"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
	"io"
	"runtime"
	"strconv"
)

type Handler[T any] struct {
	env    *HandleEnv
	data   any
	info   error
	invoke func(*Context[T]) error
}

func (h *Handler[T]) Data() any {
	return h.data
}

func (h *Handler[T]) Writer(w io.Writer, c *Context[T]) error {
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
		text, err := MarshalText(item)
		if err != nil {
			errs.Try(strconv.Itoa(i), err)
			continue
		}
		handle(i, text)
	}

	return errs.Wrap()
}

func (h *Handler[T]) SafeCall(fn func(v T) error, ctx *Context[T]) error {
	if ctx.size == 0 {
		return nil
	}

	errs := &errkit.JoinError{}
	for i := 0; i < ctx.size; i++ {
		item := ctx.data[i]
		errs.Try(strconv.Itoa(i), fn(item))
	}
	return errs.Wrap()
}

func (h *Handler[T]) Protect() {
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

func (h *Handler[T]) LFunc(fn *lua.LFunction, ctx *Context[T]) error {
	defer h.Protect()
	err := h.pcall(fn, ctx)
	return err
}

func (h *Handler[T]) pcall(fn *lua.LFunction, ctx *Context[T]) error {
	if h.env == nil {
		return fmt.Errorf("pipe pcall env is nil")
	}

	cp := lua.P{
		Protect: h.env.Protect,
		NRet:    0,
		Fn:      fn,
	}

	co := h.env.Parent.Coroutine()
	defer func() {
		h.env.Parent.Keepalive(co)
	}()

	sz := len(ctx.data)
	param := make([]lua.LValue, sz)
	for i := 0; i < sz; i++ {
		item := ctx.data[i]
		param[i] = deflect.ToLValueL(co, item)
	}

	err := co.CallByParam(cp, param...)
	return err
}

func (h *Handler[T]) Invoke(a *Context[T]) error {
	if h.invoke == nil {
		return h.info
	}
	return h.invoke(a)
}

func (h *Handler[T]) InvokerFunc(v any) {
	switch elem := v.(type) {
	case io.Writer:
		h.invoke = func(a *Context[T]) error {
			return h.Writer(elem, a)
		}
	case Invoker[T]:
		h.invoke = func(a *Context[T]) error {
			return h.SafeCall(elem.Invoke, a)
		}

	case func(T):
		h.invoke = func(a *Context[T]) error {
			defer h.Protect()
			elem(a.first())
			return nil
		}
	case func(T) error:
		h.invoke = func(a *Context[T]) error {
			defer h.Protect()
			return elem(a.first())
		}

	case *lua.LFunction:
		h.invoke = func(a *Context[T]) error {
			return h.LFunc(elem, a)
		}

	case lua.GoFuncErr:
		h.invoke = func(a *Context[T]) error {
			defer h.Protect()
			return elem(a.Use()...)
		}

	case lua.GoFuncStr:
		h.invoke = func(a *Context[T]) error {
			defer h.Protect()
			_ = elem(a.Use()...)
			return nil
		}
	case lua.GoFuncInt:
		h.invoke = func(a *Context[T]) error {
			defer h.Protect()
			_ = elem(a.Use()...)
			return nil
		}
	case lua.GoFunction[T]:
		h.invoke = func(a *Context[T]) error {
			defer h.Protect()
			elem(a.first())
			return nil
		}

	case *lua.LUserData:
		h.InvokerFunc(elem.Value)
	case lua.GenericType:
		h.InvokerFunc(elem.UnwrapData())
	case lua.Invoker:
		h.invoke = func(a *Context[T]) error {
			return h.SafeCall(func(v T) error { return elem(v) }, a)
		}
	case func(any):
		h.invoke = func(a *Context[T]) error {
			return h.SafeCall(func(v T) error { elem(v); return nil }, a)
		}
	case func(any) error:
		h.invoke = func(a *Context[T]) error {
			return h.SafeCall(func(v T) error { return elem(v) }, a)
		}
	default:
		h.info = fmt.Errorf("not compatible %T", v)
	}

}

func (h *Handler[T]) prepare(v any) {
	h.data = v
	h.InvokerFunc(v)
}
