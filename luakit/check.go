package luakit

import (
	"github.com/vela-public/onekit/lua"
)

func Check[T any](L *lua.LState, val lua.LValue) (t T) {
	if val.Type() != lua.LTObject {
		L.RaiseError("must object type , got:%s", val.Type().String())
		return
	}

	if v, ok := val.(T); !ok {
		L.RaiseError("must %T got %T", t, v)
		return v
	}

	return
}
