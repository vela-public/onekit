package luakit

import "github.com/vela-public/onekit/lua"

func NewLuaState(opts ...lua.Options) *lua.LState {
	co := lua.NewState(opts...)
	co.PreloadModule("luakit", PreloadModule)
	return co
}
