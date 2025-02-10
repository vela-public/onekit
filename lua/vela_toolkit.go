package lua

import (
	"errors"
	"net"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

var (
	InvalidFormat = errors.New("invalid format")
	InvalidIP     = errors.New("invalid ip addr")
	InvalidPort   = errors.New("expect check socket err: port <1 or port > 65535")
)

func IsString(v LValue) string {
	d, ok := v.AssertString()
	if ok {
		return d
	}

	return ""
}

func IsTrue(v LValue) bool {
	if lv, ok := v.(LBool); ok {
		return bool(lv) == true
	}
	return false
}

func IsFalse(v LValue) bool {
	if lv, ok := v.(LBool); ok {
		return bool(lv) == false
	}
	return false
}

func IsNumber(v LValue) LNumber {
	if lv, ok := v.(LNumber); ok {
		return lv
	}
	return 0
}

func IsInt(v LValue) int {
	if intv, ok := v.(LNumber); ok {
		return int(intv)
	}

	if intv, ok := v.(LInt); ok {
		return int(intv)
	}

	return 0
}

func IsFunc(v LValue) *LFunction {
	fn, _ := v.(*LFunction)
	return fn
}

func IsNull(v []byte) bool {
	if len(v) == 0 {
		return true
	}
	return false
}

func CheckInt(L *LState, lv LValue) int {
	if intv, ok := lv.(LNumber); ok {
		return int(intv)
	}
	L.RaiseError("must be int , got %s", lv.Type().String())
	return 0
}

func CheckIntOrDefault(L *LState, lv LValue, d int) int {
	if intv, ok := lv.(LNumber); ok {
		return int(intv)
	}
	return d
}

func CheckInt64(L *LState, lv LValue) int64 {
	if intv, ok := lv.(LNumber); ok {
		return int64(intv)
	}
	L.RaiseError("must be int64 , got %s", lv.Type().String())
	return 0
}

func CheckNumber(L *LState, lv LValue) LNumber {
	if lv, ok := lv.(LNumber); ok {
		return lv
	}
	L.RaiseError("must be LNumber , got %s", lv.Type().String())
	return 0
}

func CheckString(L *LState, lv LValue) string {
	if lv, ok := lv.(LString); ok {
		return string(lv)
	} else if LVCanConvToString(lv) {
		return LVAsString(lv)
	}
	return ""
}

func CheckBool(L *LState, lv LValue) bool {
	if lv, ok := lv.(LBool); ok {
		return bool(lv)
	}

	L.RaiseError("must be bool , got %s", lv.Type().String())
	return false
}

func CheckTable(L *LState, lv LValue) *LTable {
	if lv, ok := lv.(*LTable); ok {
		return lv
	}
	L.RaiseError("must be LTable, got %s", lv.Type().String())
	return nil
}

func CheckFunction(L *LState, lv LValue) *LFunction {
	if lv, ok := lv.(*LFunction); ok {
		return lv
	}
	L.RaiseError("must be Function, got %s", lv.Type().String())
	return nil
}
func CheckSocket(v string) error {
	s := strings.Split(v, ":")
	if len(s) != 2 {
		return InvalidFormat
	}

	if net.ParseIP(s[0]) == nil {
		return InvalidIP
	}

	port, err := strconv.Atoi(s[1])
	if err != nil {
		return err
	}
	if port < 1 || port > 65535 {
		return InvalidPort
	}
	return nil
}

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

func S2L(s string) LString {
	return LString(s)
}

func B2L(b []byte) LString {
	return *(*LString)(unsafe.Pointer(&b))
}

func MustBe[T any](L *LState, idx int) T {
	lv := L.Get(idx)
	vt, ok := lv.(T)
	if ok {
		return vt
	}

	L.RaiseError("must be %T , got %s", vt, lv.Type().String())
	return vt
}

func UnPack[T any](L *LState) []T {
	n := L.GetTop()
	if n == 0 {
		return nil
	}
	var rc []T
	for i := 1; i <= n; i++ {
		rc = append(rc, MustBe[T](L, i))
	}
	return rc
}

func UnpackGo(L *LState) []any {
	n := L.GetTop()
	if n == 0 {
		return nil
	}

	var rc []any
	for i := 1; i <= n; i++ {
		lv := L.Get(i)
		switch lv.Type() {
		case LTString:
			rc = append(rc, lv.String())
		case LTNumber:
			rc = append(rc, float64(lv.(LNumber)))
		case LTBool:
			if lv.(LBool) == true {
				rc = append(rc, true)
			} else {
				rc = append(rc, false)
			}
		case LTNil:
			rc = append(rc, nil)
		case LTFunction:
			rc = append(rc, lv.(*LFunction))
		default:
			rc = append(rc, lv)
		}
	}
	return rc
}

func L2SS(L *LState) []string {
	n := L.GetTop()
	if n == 0 {
		return nil
	}

	var ssv []string
	for i := 1; i <= n; i++ {
		lv := L.Get(i)
		if lv.Type() == LTNil {
			continue
		}
		v := lv.String()
		if len(v) == 0 {
			continue
		}
		ssv = append(ssv, v)
	}
	return ssv
}

func NewFunction(gn LGFunction) *LFunction {
	return &LFunction{
		IsG:       true,
		Proto:     nil,
		GFunction: gn,
	}
}

func CreateTable(acap, hcap int) *LTable {
	return newLTable(acap, hcap)
}

func CloneTable(v *LTable) *LTable {
	tab := &LTable{
		Metatable: v.Metatable,
	}

	a := len(v.array)
	if a > 0 {
		tab.array = make([]LValue, a)
		for i := 0; i < a; i++ {
			tab.array[i] = v.array[i]
		}
	}

	d := len(v.dict)
	if d > 0 {
		tab.dict = make(map[LValue]LValue, d)
		for key, val := range v.dict {
			tab.dict[key] = val
		}
	}
	s := len(v.strdict)
	if s > 0 {
		tab.strdict = make(map[string]LValue, d)
		for key, val := range v.strdict {
			tab.strdict[key] = val
		}
	}

	k := len(v.keys)
	if k > 0 {
		tab.array = make([]LValue, k)
		for i := 0; i < k; i++ {
			tab.keys[i] = v.keys[i]
		}
	}

	ki := len(v.k2i)
	if ki > 0 {
		tab.k2i = make(map[LValue]int, ki)
		for key, val := range v.k2i {
			tab.k2i[key] = val
		}
	}

	return tab
}
