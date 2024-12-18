package wasm

import (
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"unsafe"
)

/*
	{tag:"offset",type:"int32",size:4},
	{tag:"age",type:"int32",size:4},
	{tag:"hash",type:"text",size:32},
	{tag:"flag",type:"bool",size:1},

*/

type BinaryMask struct {
	Tag  string `lua:"tag"`
	Typ  string `lua:"type"`
	Size int    `lua:"size"`
}

func (b BinaryMask) Convert(text []byte) lua.LValue {
	if sz := len(text); sz != b.Size {
		return lua.LNil
	}

	switch b.Typ {
	case "int32":
		return lua.LInt(*(*int32)(unsafe.Pointer(&text[0])))
	case "uint32":
		return lua.LUint(*(*uint32)(unsafe.Pointer(&text[0])))
	case "int64":
		return lua.LInt64(*(*uint64)(unsafe.Pointer(&text[0])))
	case "uint64":
		return lua.LUint64(*(*uint64)(unsafe.Pointer(&text[0])))
	case "bool":
		return lua.LBool(text[0] != 0)
	case "text":
		return lua.B2L(text)
	}
	return lua.LNil
}

type BinaryType struct {
	mask []BinaryMask
	data []byte
}

func (bt *BinaryType) String() string                    { return cast.B2S(bt.data) }
func (bt *BinaryType) Type() lua.LValueType              { return lua.LTObject }
func (bt *BinaryType) AssertFloat64() (float64, bool)    { return 0, false }
func (bt *BinaryType) AssertString() (string, bool)      { return bt.String(), true }
func (bt *BinaryType) Hijack(fsm *lua.CallFrameFSM) bool { return false }
func (bt *BinaryType) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(func(L *lua.LState) int {
		tab := L.CheckTable(1)
		_ = luakit.TableTo(L, tab, &bt.mask)
		return 0
	}), true
}

func (bt *BinaryType) Index(L *lua.LState, key string) lua.LValue {
	sz := len(bt.mask)
	if sz == 0 {
		return lua.LNil
	}

	offset := 0
	for i := 0; i < sz; i++ {
		mask := bt.mask[i]
		if mask.Tag == key {
			return mask.Convert(bt.data[offset : offset+mask.Size])
		}
		offset += mask.Size
	}
	return lua.LNil
}
