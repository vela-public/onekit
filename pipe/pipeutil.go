package pipe

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/deflect"
	"github.com/vela-public/onekit/lua"

	"io"
)

type PCall interface {
	PCall(v ...interface{}) error
}

func (px *Chains) Len() int {
	return len(px.chain)
}

func (px *Chains) LValue(lv lua.LValue) {
	switch lv.Type() {

	case lua.LTUserData:
		px.Prepare(lv.(*lua.LUserData).Value)
	case lua.LTVelaData:
		px.Prepare(lv.(*lua.VelaData).Data)
	case lua.LTObject:
		px.Prepare(lv)
	case lua.LTGoFuncErr:
		px.LFuncErr(lv.(lua.GoFuncErr))
	case lua.LTGoFuncStr:
		px.LFuncStr(lv.(lua.GoFuncStr))
	case lua.LTGoFuncInt:
		px.LFuncInt(lv.(lua.GoFuncInt))
	case lua.LTGoFunction:
		px.GoFunc(lv.(lua.GoFunction))
	case lua.LTFunction:
		px.LFunc(lv.(*lua.LFunction))

	default:
		px.invalid("invalid pipe lua type , got %s", lv.Type().String())
	}
}

func (px *Chains) Prepare(obj interface{}) {
	switch value := obj.(type) {
	case lua.LValue:
		px.LValue(value)

	case io.Writer:
		px.Writer(value)

	case lua.Console:
		px.append(px.Console(value))

	case PCall:
		px.append(value.PCall)

	case func():
		px.append(func(...interface{}) error {
			value()
			return nil
		})

	case func(interface{}):
		px.append(func(v ...interface{}) error {
			if len(v) == 0 {
				value(nil)
			} else {
				value(v[0])
			}
			return nil
		})

	case func() error:
		px.append(func(...interface{}) error {
			_ = value()
			return nil
		})

	case func(interface{}) error:
		px.append(func(v ...interface{}) error {
			if len(v) == 0 {
				return value(nil)
			} else {
				return value(v[0])
			}
		})

	default:
		px.invalid("invalid pipe object")
		return
	}
}

func (px *Chains) GoFunc(fn lua.GoFunction) {
	px.append(func(v ...interface{}) error {
		return fn()
	})
}

func (px *Chains) LFuncErr(fn lua.GoFuncErr) {
	px.append(func(v ...interface{}) error {
		return fn(v...)
	})
}

func (px *Chains) LFuncStr(fn lua.GoFuncStr) {
	px.append(func(v ...interface{}) error {
		fn(v...)
		return nil
	})
}

func (px *Chains) LFuncInt(fn lua.GoFuncInt) {
	px.append(func(v ...interface{}) error {
		fn(v...)
		return nil
	})
}

func (px *Chains) LFunc(fn *lua.LFunction) {
	px.append(func(v ...interface{}) error {
		size := len(v)
		if size == 0 {
			return nil
		}

		var co *lua.LState
		L, ok := v[size-1].(*lua.LState)
		if ok {
			co = px.clone(L)
			size = size - 1
		}
		cp := lua.P{
			Protect: true,
			NRet:    0,
			Fn:      fn,
		}

		args := make([]lua.LValue, size)
		for i := 0; i < size; i++ {
			args[i] = deflect.ToLValueL(co, v[i])
		}
		defer px.xEnv.Free(co)

		if len(args) == 0 {
			return fmt.Errorf("reflectx to LValue fail %v", v)
		}

		return co.CallByParam(cp, args...)
	})
}

func (px *Chains) write(w io.Writer, v ...interface{}) error {
	size := len(v)
	if size == 0 {
		return nil
	}

	data, err := cast.ToStringE(v[0])
	if err != nil {
		return err
	}
	_, err = w.Write(cast.S2B(data))
	return err
}

func (px *Chains) Writer(w io.Writer) {
	px.append(func(v ...interface{}) error {
		if w == nil {
			return fmt.Errorf("invalid io writer %p", w)
		}

		return px.write(w, v...)
	})
}

func (px *Chains) SetEnv(env Environment) *Chains {
	if env != nil {
		px.xEnv = env
	}
	return px
}

func (px *Chains) Console(out lua.Console) Handler {
	return func(v ...interface{}) error {
		size := len(v)
		if size == 0 {
			return nil
		}

		data, err := cast.ToStringE(v[0])
		if err != nil {
			return err
		}
		out.Println(data)
		return nil
	}
}

func (px *Chains) InvokeGo(v interface{}, x func(error) (stop bool)) {
	sz := len(px.chain)
	if sz == 0 {
		return
	}

	for i := 0; i < sz; i++ {
		fn := px.chain[i]
		if e := fn(v); e != nil && x != nil {
			if x(e) {
				return
			}
		}
	}
}

func (px *Chains) Invoke(v interface{}, x func(error) (stop bool)) {
	n := len(px.chain)
	if n == 0 {
		return
	}

	co := px.xEnv.Clone(px.vm)
	defer px.xEnv.Free(co)

	for i := 0; i < n; i++ {
		fn := px.chain[i]
		if e := fn(v, co); e != nil && x != nil {
			if x(e) {
				return
			}
		}
	}

}

//兼容老的数据

func (px *Chains) Do(arg interface{}, co *lua.LState, x func(error)) {
	n := len(px.chain)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		fn := px.chain[i]
		if e := fn(arg, co); e != nil && x != nil {
			x(e)
		}
	}
}

func (px *Chains) Case(v interface{}, id int, cnd string, co *lua.LState) error {
	n := len(px.chain)
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		fn := px.chain[i]
		if e := fn(v, id, cnd, co); e != nil {
			return e
		}
	}

	return nil
}

func (px *Chains) Call2(v1 interface{}, v2 interface{}, co *lua.LState) error {
	n := len(px.chain)
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		fn := px.chain[i]
		if e := fn(v1, v2, co); e != nil {
			return e
		}
	}

	return nil
}

func (px *Chains) Call(co *lua.LState, v ...interface{}) error {
	n := len(px.chain)
	if n == 0 {
		return nil
	}

	param := append(v, co)
	for i := 0; i < n; i++ {
		fn := px.chain[i]
		if e := fn(param...); e != nil {
			return e
		}
	}

	return nil
}
