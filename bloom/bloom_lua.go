package bloom

import (
	"encoding/json"
	"github.com/vela-public/onekit/lua"
)

func (bf *Filter) String() string {
	text, _ := json.Marshal(bf)
	return lua.B2S(text)
}

func (bf *Filter) Type() lua.LValueType                   { return lua.LTObject }
func (bf *Filter) AssertFloat64() (float64, bool)         { return 0, false }
func (bf *Filter) AssertString() (string, bool)           { return "", false }
func (bf *Filter) AssertFunction() (*lua.LFunction, bool) { return lua.NewFunction(bf.upsertL), true }
func (bf *Filter) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (bf *Filter) upsertL(L *lua.LState) int {
	top := L.GetTop()
	if top <= 0 {
		return 0
	}

	for i := 1; i <= top; i++ {
		item := L.CheckString(i)
		bf.Upsert(item)
	}

	L.Push(bf)
	return 1
}

func (bf *Filter) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "upsert":
		return lua.NewFunction(bf.upsertL)
	case "size":
		return lua.LUint(bf.Size)
	case "hashes":
		return lua.LInt(bf.Hashes)
	case "cnt":
		return lua.LInt(bf.Cnt)
	}

	return lua.LNil
}
