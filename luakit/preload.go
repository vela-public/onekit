package luakit

import (
	"github.com/vela-public/onekit/lua"
)

func builtin(kv lua.UserKV) {
	kv.Set("fmt", lua.NewFunction(NewFmtL))
}
