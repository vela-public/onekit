package lua

import (
	"github.com/vela-public/onekit/cast"
	"strconv"
)

const (
	IntegerLibName = "int"
)

type LInt int
type LUint uint
type LInt64 int64
type LUint64 uint64

func (i LInt) Type() LValueType                   { return LTInt }
func (i LInt) String() string                     { return strconv.Itoa(int(i)) }
func (i LInt) AssertFloat64() (float64, bool)     { return float64(i), true }
func (i LInt) AssertString() (string, bool)       { return "", false }
func (i LInt) AssertFunction() (*LFunction, bool) { return nil, false }
func (i LInt) Hijack(*CallFrameFSM) bool          { return false }

func (ui LUint) Type() LValueType                   { return LTUint }
func (ui LUint) String() string                     { return cast.ToString(uint32(ui)) }
func (ui LUint) AssertFloat64() (float64, bool)     { return cast.ToFloat64(uint(ui)), true }
func (ui LUint) AssertString() (string, bool)       { return "", false }
func (ui LUint) AssertFunction() (*LFunction, bool) { return nil, false }
func (ui LUint) Hijack(*CallFrameFSM) bool          { return false }

func (i LInt64) Type() LValueType                   { return LTInt64 }
func (i LInt64) String() string                     { return strconv.Itoa(int(i)) }
func (i LInt64) AssertFloat64() (float64, bool)     { return float64(i), true }
func (i LInt64) AssertString() (string, bool)       { return "", false }
func (i LInt64) AssertFunction() (*LFunction, bool) { return nil, false }
func (i LInt64) Hijack(*CallFrameFSM) bool          { return false }

func (ui LUint64) Type() LValueType                   { return LTUint64 }
func (ui LUint64) String() string                     { return strconv.FormatUint(uint64(ui), 10) }
func (ui LUint64) AssertFloat64() (float64, bool)     { return float64(ui), true }
func (ui LUint64) AssertString() (string, bool)       { return "", false }
func (ui LUint64) AssertFunction() (*LFunction, bool) { return nil, false }
func (ui LUint64) Hijack(*CallFrameFSM) bool          { return false }

func OpenIntegerLib(L *LState) int {
	L.SetGlobal("int", NewFunction(func(c *LState) int {
		lv := L.Get(1)
		switch lv.Type() {
		case LTNumber:
			L.Push(LInt(lv.(LNumber)))
		case LTInt:
			L.Push(LInt(lv.(LInt)))
		case LTInt64:
			L.Push(LInt(lv.(LInt64)))
		case LTUint:
			L.Push(LInt(lv.(LUint)))
		case LTUint64:
			L.Push(LInt(lv.(LUint64)))
		default:
			L.Push(LInt(0))
		}
		return 1
	}))
	L.SetGlobal("int64", NewFunction(func(c *LState) int {
		lv := L.Get(1)
		switch lv.Type() {
		case LTNumber:
			L.Push(LInt64(lv.(LNumber)))
		case LTInt:
			L.Push(LInt64(lv.(LInt)))
		case LTInt64:
			L.Push(LInt64(lv.(LInt64)))
		case LTUint:
			L.Push(LInt64(lv.(LUint)))
		case LTUint64:
			L.Push(LInt64(lv.(LUint64)))
		default:
			L.Push(LInt64(0))
		}
		return 1
	}))
	return 0
}
