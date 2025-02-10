package luakit

import (
	"github.com/vela-public/onekit/lua"
)

func builtin(kv lua.UserKV) {
	kv.Set("pretty", lua.NewFunction(PrettyJsonL))
	kv.Set("fmt", lua.NewFunction(NewFmtL))
	kv.Set("pretty", lua.NewFunction(PrettyJsonL))
}
