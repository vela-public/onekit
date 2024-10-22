package timekit

import (
	"github.com/vela-public/onekit/lua"
	"strconv"
	"strings"
	"time"
)

func subL(L *lua.LState) int {
	layout := func() string {
		d := L.IsString(3)
		if d == "" {
			d = "2006-01-02 15:04:05"
		}
		return d
	}

	v1 := UnixL(L, L.Get(1), layout())
	v2 := UnixL(L, L.Get(2), layout())
	if v1 < 0 || v2 < 0 {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(lua.LNumber(v1 - v2))
	return 1
}

// time.scope(t , 5 , "2006")
func scopeL(L *lua.LState) int {
	layout := func() string {
		d := L.IsString(3)
		if d == "" {
			d = "2006-01-02 15:04:05"
		}
		return d
	}

	v := UnixL(L, L.Get(1), layout())
	s := L.Get(2)

	sn := int64(0)
	switch s.Type() {
	case lua.LTString:
		switch s.String() {
		case "day":
			sn = 86400
		case "hour":
			sn = 3600
		case "min":
			sn = 60
		default:
			L.RaiseError(s.String() + " Not a valid scope description")
			return 0
		}
	case lua.LTNumber:
		sn = int64(s.(lua.LNumber))
	case lua.LTInt:
		sn = int64(s.(lua.LInt))
	case lua.LTInt64:
		sn = int64(s.(lua.LInt64))
	}

	if sn == 0 {
		L.Push(lua.LInt64(v))
		return 1
	}

	remainder := v % sn
	L.Push(lua.LInt64(v - remainder))
	return 1
}

func HourL(L *lua.LState) int {
	L.Push(lua.LInt(time.Now().Hour()))
	return 1
}

// ScopeHourL time.scope_hour("9-12" , "13-15")
func ScopeHourL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		L.Push(lua.LFalse)
		return 1
	}

	hour := time.Now().Hour()

	for idx := 1; idx <= n; idx++ {
		lv := L.Get(idx)
		switch lv.Type() {
		case lua.LTNumber:
			if hour == int(lv.(lua.LNumber)) {
				L.Push(lua.LTrue)
				return 1
			}
		case lua.LTInt:
			if hour == int(lv.(lua.LInt)) {
				L.Push(lua.LTrue)
				return 1
			}
		case lua.LTString:
			val := lv.String()
			if v, e := strconv.Atoi(val); e == nil {
				if v == hour {
					L.Push(lua.LTrue)
					return 1
				}
				continue
			}

			if r := strings.Split(val, "-"); len(r) == 2 {
				var a, b int
				var e error
				a, e = strconv.Atoi(r[0])
				b, e = strconv.Atoi(r[1])
				if e != nil {
					continue
				}

				if hour >= a && b >= hour {
					L.Push(lua.LTrue)
					return 1
				}

			}

		default:
			//todo
		}

	}

	L.Push(lua.LFalse)
	return 1
}
