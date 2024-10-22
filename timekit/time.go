package timekit

import (
	"github.com/vela-public/onekit/lua"
	"time"
)

func indexL(L *lua.LState, key string) lua.LValue {
	switch key {
	case "now":
		return lua.LNumber(time.Now().Unix())
	case "sub":
		return lua.NewFunction(subL)
	case "scope":
		return lua.NewFunction(scopeL)
	case "hour":
		return lua.NewFunction(HourL)
	case "scope_hour":
		return lua.NewFunction(ScopeHourL)
	}

	return lua.LNil
}

func timeL(L *lua.LState) int {
	v := L.Get(1)
	layout := func() string {
		d := L.IsString(2)
		if d == "" {
			d = "2006-01-02 15:04:05"
		}
		return d
	}

	switch v.Type() {
	case lua.LTNumber:
		t := time.Unix(int64(v.(lua.LNumber)), 0)
		L.Push(lua.S2L(t.Format(layout())))
		return 1
	case lua.LTInt64:
		t := time.Unix(int64(v.(lua.LInt64)), 0)
		L.Push(lua.S2L(t.Format(layout())))
		return 1
	case lua.LTString:
		t, err := time.Parse(layout(), v.String())
		if err != nil {
			//L.RaiseError("time parse fail %v", err)
			L.Push(lua.LInt64(0))
			L.Push(lua.S2L(err.Error()))
			return 2
		}
		L.Push(lua.LInt64(t.Unix()))
		return 1
	case lua.LTNil:
		L.Push(lua.LInt64(time.Now().Unix()))
		return 1
	case lua.LTBool:
		L.Push(lua.S2L(time.Now().Format(layout())))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func Preload(v lua.Preloader) {
	v.Set("time", lua.NewExport("lua.time.export", lua.WithIndex(indexL), lua.WithFunc(timeL)))
}
