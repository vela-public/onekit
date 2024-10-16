package pipe

/*
func (ch *Chains) Len() int {
	return len(ch.chain)
}

func (ch *Chains) LValue(lv lua.LValue) {
	switch lv.Type() {

	case lua.LTUserData:
		ch.Prepare(lv.(*lua.LUserData).Value)
	case lua.LTVelaData:
		ch.Prepare(lv.(*lua.VelaData).Data)
	case lua.LTObject:
		ch.Prepare(lv)
	case lua.LTGoFuncErr:
		ch.LFuncErr(lv.(lua.GoFuncErr))
	case lua.LTGoFuncStr:
		ch.LFuncStr(lv.(lua.GoFuncStr))
	case lua.LTGoFuncInt:
		ch.LFuncInt(lv.(lua.GoFuncInt))
	case lua.LTGoFunction:
		ch.GoFunc(lv.(lua.GoFunction))
	case lua.LTFunction:
		ch.LFunc(lv.(*lua.LFunction))

	default:
		ch.invalid("invalid pipe pool type , got %s", lv.Type().String())
	}
}

func (ch *Chains) Prepare(obj interface{}) {
	switch value := obj.(type) {
	case lua.LValue:
		ch.LValue(value)

	case io.Single:
		ch.Single(value)

	case lua.Console:
		ch.append(ch.Console(value))

	case func():
		ch.append(func(...interface{}) error {
			value()
			return nil
		})

	case func(interface{}):
		ch.append(func(v ...interface{}) error {
			if len(v) == 0 {
				value(nil)
			} else {
				value(v[0])
			}
			return nil
		})

	case func() error:
		ch.append(func(...interface{}) error {
			_ = value()
			return nil
		})

	case func(interface{}) error:
		ch.append(func(v ...interface{}) error {
			if len(v) == 0 {
				return value(nil)
			} else {
				return value(v[0])
			}
		})

	default:
		ch.invalid("invalid pipe object")
		return
	}
}

func (ch *Chains) GoFunc(fn lua.GoFunction) {
	ch.append(func(v ...interface{}) error {
		return fn()
	})
}

func (ch *Chains) LFuncErr(fn lua.GoFuncErr) {
	ch.append(func(v ...interface{}) error {
		return fn(v...)
	})
}

func (ch *Chains) LFuncStr(fn lua.GoFuncStr) {
	ch.append(func(v ...interface{}) error {
		fn(v...)
		return nil
	})
}

func (ch *Chains) LFuncInt(fn lua.GoFuncInt) {
	ch.append(func(v ...interface{}) error {
		fn(v...)
		return nil
	})
}

func (ch *Chains) LFunc(fn *lua.LFunction) {
	ch.append(func(v ...interface{}) error {
		size := len(v)
		if size == 0 {
			return nil
		}

		var co *lua.Main
		L, ok := v[size-1].(*lua.Main)
		if ok {
			co = ch.clone(L)
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
		defer ch.xEnv.Free(co)

		if len(args) == 0 {
			return fmt.Errorf("reflectx to LValue fail %v", v)
		}

		return co.CallByParam(cp, args...)
	})
}

func (ch *Chains) write(w io.Single, v ...interface{}) error {
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

func (ch *Chains) Single(w io.Single) {
	ch.append(func(v ...interface{}) error {
		if w == nil {
			return fmt.Errorf("invalid io writer %p", w)
		}

		return ch.write(w, v...)
	})
}

func (ch *Chains) SetEnv(env LuaVM) *Chains {
	if env != nil {
		ch.xEnv = env
	}
	return ch
}

func (ch *Chains) Console(out lua.Console) Handler {
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

func (ch *Chains) InvokeGo(v interface{}, x func(error) (stop bool)) {
	sz := len(ch.chain)
	if sz == 0 {
		return
	}

	for i := 0; i < sz; i++ {
		fn := ch.chain[i]
		if e := fn(v); e != nil && x != nil {
			if x(e) {
				return
			}
		}
	}
}

func (ch *Chains) Invoke(v interface{}, x func(error) (stop bool)) {
	n := len(ch.chain)
	if n == 0 {
		return
	}

	co := ch.xEnv.Clone(ch.vm)
	defer ch.xEnv.Free(co)

	for i := 0; i < n; i++ {
		fn := ch.chain[i]
		if e := fn(v, co); e != nil && x != nil {
			if x(e) {
				return
			}
		}
	}

}

//兼容老的数据

func (ch *Chains) Do(arg interface{}, co *lua.Main, x func(error)) {
	n := len(ch.chain)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		fn := ch.chain[i]
		if e := fn(arg, co); e != nil && x != nil {
			x(e)
		}
	}
}

func (ch *Chains) Case(v interface{}, id int, cnd string, co *lua.Main) error {
	n := len(ch.chain)
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		fn := ch.chain[i]
		if e := fn(v, id, cnd, co); e != nil {
			return e
		}
	}

	return nil
}

func (ch *Chains) Call2(v1 interface{}, v2 interface{}, co *lua.Main) error {
	n := len(ch.chain)
	if n == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		fn := ch.chain[i]
		if e := fn(v1, v2, co); e != nil {
			return e
		}
	}

	return nil
}

func (ch *Chains) Call(co *lua.Main, v ...interface{}) error {
	n := len(ch.chain)
	if n == 0 {
		return nil
	}

	param := append(v, co)
	for i := 0; i < n; i++ {
		fn := ch.chain[i]
		if e := fn(param...); e != nil {
			return e
		}
	}

	return nil
}
*/
