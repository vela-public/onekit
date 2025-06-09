package abi

import "github.com/vela-public/onekit/lua"

func NewStructInstanceL(L *lua.LState) int {
	packed := L.IsTrue(1)
	builder := NewStructBuilder(packed)
	L.Push(builder)
	return 1
}
