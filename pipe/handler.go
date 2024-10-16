package pipe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vela-public/onekit/cast"
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
		switch v := item.(type) {
		case string:
			handle(i, cast.S2B(v))
		case []byte:
			handle(i, v)
		case fmt.Stringer:
			handle(i, cast.S2B(v.String()))
		case bytes.Buffer:
			handle(i, v.Bytes())
		case *bytes.Buffer:
			handle(i, v.Bytes())
		case Bytes:
			handle(i, v.Bytes())
		case int8:
			text := strconv.FormatInt(int64(v), 10)
			handle(i, cast.S2B(text))
		case int16:
			text := strconv.FormatInt(int64(v), 10)
			handle(i, cast.S2B(text))
		case int32:
			text := strconv.FormatInt(int64(v), 10)
			handle(i, cast.S2B(text))
		case int:
			text := strconv.FormatInt(int64(v), 10)
			handle(i, cast.S2B(text))
		case int64:
			text := strconv.FormatInt(v, 10)
			handle(i, cast.S2B(text))
		case uint8:
			text := strconv.FormatUint(uint64(v), 10)
			handle(i, cast.S2B(text))
		case uint16:
			text := strconv.FormatUint(uint64(v), 10)
			handle(i, cast.S2B(text))
		case uint32:
			text := strconv.FormatUint(uint64(v), 10)
			handle(i, cast.S2B(text))
		case uint:
			text := strconv.FormatUint(uint64(v), 10)
			handle(i, cast.S2B(text))
		case uint64:
			text := strconv.FormatUint(uint64(v), 10)
			handle(i, cast.S2B(text))

		case float32:
			text := strconv.FormatFloat(float64(v), 'f', -1, 64)
			handle(i, cast.S2B(text))
		case float64:
			text := strconv.FormatFloat(v, 'f', -1, 64)
			handle(i, cast.S2B(text))
		case bool:
			text := strconv.FormatBool(v)
			handle(i, cast.S2B(text))

		default:
			text, err := json.Marshal(item)
			if err != nil {
				errs.Try(strconv.Itoa(i), err)
			} else {
				handle(i, text)
			}
		}
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
	case lua.GoFunction:
		h.invoke = func(a *Context) error {
			defer h.Protect()
			return elem()
		}
	case *lua.LUserData:
		InvokerFunc(h, elem.Value)
	case *lua.VelaData:
		InvokerFunc(h, elem.Data)
	case func(any):
		h.invoke = func(a *Context) error {
			return h.SafeCall(func(v any) error { elem(v); return nil }, a)
		}
	default:
		h.info = fmt.Errorf("not compatible %T", v)
	}

}

func (h *Handler) prepare(v any) {
	h.data = v
	InvokerFunc(h, v)
}
