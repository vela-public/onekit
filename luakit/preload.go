package luakit

import (
	"github.com/vela-public/onekit/lua"
)

func builtin(kv lua.UserKV) {
	kv.Set("map", lua.NewFunction(NewMapL))
	kv.Set("slice", lua.NewFunction(NewSliceL))
	kv.Set("int", lua.NewFunction(NewIntL))
	kv.Set("pretty", lua.NewFunction(PrettyJsonL))
	kv.Set("fmt", lua.NewFunction(NewFmtL))
	kv.Set("pretty", lua.NewFunction(PrettyJsonL))
}
