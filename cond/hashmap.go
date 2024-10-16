package cond

import (
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"path/filepath"
)

type Hashmap struct {
	name string
	data map[string]libkit.NULL
}

func (h *Hashmap) String() string                         { return "hashmap.filter." + h.name }
func (h *Hashmap) Type() lua.LValueType                   { return lua.LTObject }
func (h *Hashmap) AssertFloat64() (float64, bool)         { return float64(len(h.data)), true }
func (h *Hashmap) AssertString() (string, bool)           { return "", false }
func (h *Hashmap) AssertFunction() (*lua.LFunction, bool) { return lua.NewFunction(h.MatchL), true }
func (h *Hashmap) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (h *Hashmap) MatchL(L *lua.LState) int {
	lv := L.CheckString(1)
	if h.data == nil {
		L.Push(lua.LFalse)
		return 1
	}

	_, ok := h.data[lv]
	if ok {
		L.Push(lua.LTrue)
		return 1
	}

	L.Push(lua.LFalse)
	return 1
}

func (h *Hashmap) fromL(L *lua.LState) int {
	path := filepath.Clean(L.CheckString(1))
	if h.data == nil {
		h.data = make(map[string]libkit.NULL)
	}

	if e := libkit.ReadlineFunc(path, func(text string) (bool, error) {
		h.data[text] = libkit.NULL{}
		return false, nil
	}); e != nil {
		L.RaiseError("readline fail %s %v", path, e)
		return 0
	}
	L.Push(h)
	return 1
}

func (h *Hashmap) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "file":
		return lua.NewFunction(h.fromL)
	case "size":
		return lua.LInt(len(h.data))
	}
	return lua.LNil
}

func NewHashMapL(L *lua.LState) int {
	name := L.CheckString(1)
	h := &Hashmap{
		name: name,
	}
	L.Push(h)
	return 1
}
