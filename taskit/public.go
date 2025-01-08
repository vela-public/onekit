package taskit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
)

func CheckName(v string) error {
	if len(v) < 2 {
		return fmt.Errorf("too short name got:%s", v)
	}

	if !IsChar(v[0]) {
		return fmt.Errorf("first char must be a-z or A-Z got:%v", string(v[0]))
	}

	n := len(v)
	for i := 1; i < n; i++ {
		ch := v[i]
		switch {
		case IsChar(ch), IsInt(ch):
			continue
		case ch == '_':
			continue
		case ch == '-':
			continue
		default:
			return fmt.Errorf("not allowed char %v", string(ch))

		}
	}
	return nil

}

func CheckTaskEx(L *lua.LState, x func(error)) *task {
	dat := L.Payload()
	if dat == nil {
		x(fmt.Errorf("not allowed in without task private layer"))
		return nil
	}

	tas, ok := dat.(*task)
	if !ok {
		x(fmt.Errorf("not found task private layer got:%T", dat))
		return nil
	}
	return tas
}

func Check[T any](L *lua.LState, srv *Service) (t T) {
	if srv.Nil() {
		L.RaiseError("not found service data")
		return
	}

	dat, ok := srv.data.(T)
	if !ok {
		L.RaiseError("mismatch service type must:%T got:%T", t, srv.data)
		return t
	}

	return dat
}

func Create(L *lua.LState, name string, typeof string) *Service {
	if err := CheckName(name); err != nil {
		L.RaiseError("%v", err)
		return nil
	}

	tas := CheckTaskEx(L, func(e error) {
		L.RaiseError("%v", e)
	})

	if tas == nil {
		return nil
	}

	return tas.Create(L, name, typeof)
}

func Start(L *lua.LState, v ServiceType, x func(e error)) {
	tas := CheckTaskEx(L, x)
	if tas == nil {
		return
	}

	tas.Do(v, x)
}
