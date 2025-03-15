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

type Number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

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

func MustBeNumber[T Number](L *LState, lv LValue) (vt T) {
	switch lv.Type() {
	case LTNumber:
		return T(lv.(LNumber))
	case LTInt:
		return T(lv.(LInt))
	case LTInt64:
		return T(lv.(LInt64))
	case LTUint:
		return T(lv.(LUint))
	case LTUint64:
		return T(lv.(LUint64))
	default:
		L.RaiseError("must be %T , got %s", vt, lv.Type().String())
		return
	}
}

func MustBe[T any](L *LState, idx int) T {
	lv := L.Get(idx)
	vt, ok := lv.(T)
	if ok {
		return vt
	}

	switch any(vt).(type) {
	case string:
		v := lv.String()
		return *(*T)(unsafe.Pointer(&v))
	case []byte:
		v := S2B(lv.String())
		return *(*T)(unsafe.Pointer(&v))
	case float64:
		n := MustBeNumber[float64](L, lv)
		return any(n).(T)
	case float32:
		n := MustBeNumber[float32](L, lv)
		return any(n).(T)
	case int:
		n := MustBeNumber[int](L, lv)
		return any(n).(T)
	case int8:
		n := MustBeNumber[int8](L, lv)
		return any(n).(T)
	case int16:
		n := MustBeNumber[int16](L, lv)
		return any(n).(T)
	case int32:
		n := MustBeNumber[int32](L, lv)
		return any(n).(T)
	case int64:
		n := MustBeNumber[int64](L, lv)
		return any(n).(T)
	case uint:
		n := MustBeNumber[uint](L, lv)
		return any(n).(T)
	case uint8:
		n := MustBeNumber[uint8](L, lv)
		return any(n).(T)
	case uint16:
		n := MustBeNumber[uint16](L, lv)
		return any(n).(T)
	case uint32:
		n := MustBeNumber[uint32](L, lv)
		return any(n).(T)
	case uint64:
		n := MustBeNumber[uint64](L, lv)
		return any(n).(T)
	case bool:
		if lv.Type() != LTBool {
			L.RaiseError("must be %T , got %s", vt, lv.Type().String())
			return vt
		}
		if lv.(LBool) == LTrue {
			return any(true).(T)
		}
		return any(false).(T)
	default:
		L.RaiseError("must be %T , got %s", vt, lv.Type().String())
		return vt
	}
}

func UnpackSeek[T any](L *LState, seek int) []T {
	n := L.GetTop()
	if n == 0 {
		return nil
	}
	var rc []T
	for i := seek; i <= n; i++ {
		rc = append(rc, MustBe[T](L, i))
	}
	return rc
}

func Unpack[T any](L *LState) []T {
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
		case LTInt:
			rc = append(rc, int(lv.(LInt)))
		case LTInt64:
			rc = append(rc, int64(lv.(LInt64)))
		case LTUint:
			rc = append(rc, uint(lv.(LUint)))
		case LTUint64:
			rc = append(rc, uint(lv.(LUint64)))
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
		case LTGeneric:
			rc = append(rc, lv.(GenericType).Unpack())
		case LTUserData:
			rc = append(rc, lv.(*LUserData).Value)
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

func Exdata[T any](L *LState) (t T, ok bool) {
	if L == nil {
		return
	}

	t, ok = L.private.Exdata.(T)
	return
}

func Exdata2[T any](L *LState) (t T, ok bool) {
	if L == nil {
		return
	}
	t, ok = L.private.Exdata2.(T)
	return
}

func SetExdata2[T any](L *LState, t T) {
	if L == nil {
		return
	}
	L.private.Exdata2 = t
}

func Check[T any](L *LState, lv LValue) (t T) {
	to := func(v any) T {
		return v.(T)
	}

	switch any(t).(type) {
	case []byte:
		return to(S2B(lv.String()))
	case string:
		return to(lv.String())
	case bool:
		if lv.Type() == LTBool {
			if lv.(LBool) == true {
				to(true)
			} else {
				to(false)
			}
		}
		return to(false)
	case int8:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(int8(n))
		case LTInt:
			return to(int8(lv.(LInt)))
		case LTInt64:
			return to(int8(lv.(LInt64)))
		case LTUint:
			return to(int8(lv.(LUint)))
		case LTUint64:
			return to(int8(lv.(LUint)))
		default:
			return to(int8(0))
		}
	case int16:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(int16(n))
		case LTInt:
			return to(int16(lv.(LInt)))
		case LTInt64:
			return to(int16(lv.(LInt64)))
		case LTUint:
			return to(int16(lv.(LUint)))
		case LTUint64:
			return to(int16(lv.(LUint)))
		default:
			return to(int16(0))
		}
	case int32:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(int32(n))
		case LTInt:
			return to(int32(lv.(LInt)))
		case LTInt64:
			return to(int32(lv.(LInt64)))
		case LTUint:
			return to(int32(lv.(LUint)))
		case LTUint64:
			return to(int32(lv.(LUint)))
		default:
			return to(int32(0))
		}
	case int:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(int(n))
		case LTInt:
			return to(int(lv.(LInt)))
		case LTInt64:
			return to(int(lv.(LInt64)))
		case LTUint:
			return to(int(lv.(LUint)))
		case LTUint64:
			return to(int(lv.(LUint)))
		default:
			return to(int(0))
		}
	case int64:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(int64(n))
		case LTInt:
			return to(int64(lv.(LInt)))
		case LTInt64:
			return to(int64(lv.(LInt64)))
		case LTUint:
			return to(int64(lv.(LUint)))
		case LTUint64:
			return to(int64(lv.(LUint)))
		default:
			return to(int64(0))
		}
	case uint8:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(uint8(n))
		case LTInt:
			return to(uint8(lv.(LInt)))
		case LTInt64:
			return to(uint8(lv.(LInt64)))
		case LTUint:
			return to(uint8(lv.(LUint)))
		case LTUint64:
			return to(uint8(lv.(LUint)))
		default:
			return to(uint8(0))
		}
	case uint16:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(uint16(n))
		case LTInt:
			return to(uint16(lv.(LInt)))
		case LTInt64:
			return to(uint16(lv.(LInt64)))
		case LTUint:
			return to(uint16(lv.(LUint)))
		case LTUint64:
			return to(uint16(lv.(LUint)))
		default:
			return to(uint16(0))
		}
	case uint32:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(uint32(n))
		case LTInt:
			return to(uint32(lv.(LInt)))
		case LTInt64:
			return to(uint32(lv.(LInt64)))
		case LTUint:
			return to(uint32(lv.(LUint)))
		case LTUint64:
			return to(uint32(lv.(LUint)))
		default:
			return to(uint32(0))
		}
	case uint:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(uint(n))
		case LTInt:
			return to(uint(lv.(LInt)))
		case LTInt64:
			return to(uint(lv.(LInt64)))
		case LTUint:
			return to(uint(lv.(LUint)))
		case LTUint64:
			return to(uint(lv.(LUint)))
		default:
			return to(uint(0))
		}
	case uint64:
		switch lv.Type() {
		case LTNumber:
			n := lv.(LNumber)
			return to(uint64(n))
		case LTInt:
			return to(uint64(lv.(LInt)))
		case LTInt64:
			return to(uint64(lv.(LInt64)))
		case LTUint:
			return to(uint64(lv.(LUint)))
		case LTUint64:
			return to(uint64(lv.(LUint)))
		default:
			return to(uint64(0))
		}
	default:
		vt, ok := lv.(T)
		if ok {
			return vt
		}

		dat, ok := lv.(*Generic[T])
		if ok {
			return dat.Unwrap()
		}

		L.RaiseError("must be %T, got %s", t, lv.Type().String())
		return
	}
}
