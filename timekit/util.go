package timekit

import (
	"github.com/vela-public/onekit/lua"
	"time"
)

func UnixL(L *lua.LState, v lua.LValue, layout string) int64 {
	switch v.Type() {
	case lua.LTNumber:
		return int64(v.(lua.LNumber))
	case lua.LTInt64:
		return int64(v.(lua.LInt64))
	case lua.LTString:
		t, err := time.Parse(layout, v.String())
		if err != nil {
			L.RaiseError("time parse fail %v", err)
			return -1
		}
		return t.Unix()
	default:
		L.RaiseError("time convert to unix fail ,got %s", v.Type().String())
		return -1
	}

}
