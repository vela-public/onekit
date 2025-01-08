package luakit

import (
	"fmt"
	"github.com/tidwall/pretty"
	"github.com/vela-public/onekit/lua"
	"reflect"
	"unsafe"
)

func S2B(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return
}

func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func S2L(s string) lua.LString {
	return lua.LString(s)
}

func B2L(b []byte) lua.LString {
	return *(*lua.LString)(unsafe.Pointer(&b))
}

func PrettyJson(lv lua.LValue) []byte {
	chunk := S2B(lv.String())
	return pretty.PrettyOptions(chunk, nil)
}

func PrettyJsonL(L *lua.LState) int {
	L.Push(B2L(PrettyJson(L.Get(1))))
	return 1
}

func Format(L *lua.LState, seek int) string {
	n := L.GetTop()
	if seek > n {
		return ""
	}

	offset := n - seek
	switch offset {
	case 0:
		return ""
	case 1:
		return L.Get(seek + 1).String()
	default:
		format := L.CheckString(seek + 1)
		var args []interface{}

		for idx := seek + 2; idx <= n; idx++ {
			lv := L.Get(idx)
			switch lv.Type() {
			case lua.LTString:
				args = append(args, lv.String())
			case lua.LTBool:
				args = append(args, bool(lv.(lua.LBool)))
			case lua.LTNumber:
				num := lv.(lua.LNumber)
				if num == lua.LNumber(int(num)) {
					args = append(args, int(num))
				} else {
					args = append(args, num)
				}

			case lua.LTInt:
				args = append(args, int(lv.(lua.LInt)))
			case lua.LTNil:
				args = append(args, nil)
			case lua.LTFunction:
				args = append(args, lv)
			case lua.LTUserData:
				args = append(args, lv.(*lua.LUserData).Value)
			case lua.LTObject:
				args = append(args, lv)
			case lua.LTGeneric:
				args = append(args, lv.(lua.WrapType).UnwrapData())
			case lua.LTService:
				em, ok := lv.(lua.WrapType)
				if ok {
					args = append(args, em.UnwrapData())
				}

			default:
				args = append(args, lv)
			}
		}

		return fmt.Sprintf(format, args...)
	}
}

func NewFmtL(L *lua.LState) int {
	L.Push(S2L(Format(L, 0)))
	return 1
}
