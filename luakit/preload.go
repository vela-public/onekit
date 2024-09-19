package luakit

import "github.com/vela-public/onekit/lua"

type Preloader interface {
	Set(string, lua.LValue)
	SetGlobal(string, lua.LValue)
}

func PreloadModule(L *lua.LState) int {
	kv := lua.NewUserKV()
	kv.Set("map", lua.NewFunction(NewMapL))
	kv.Set("slice", lua.NewFunction(NewSliceL))
	kv.Set("int", lua.NewFunction(NewIntL))
	kv.Set("pretty", lua.NewFunction(PrettyJsonL))
	kv.Set("fmt", lua.NewFunction(NewFmtL))
	kv.Set("pretty", lua.NewFunction(PrettyJsonL))
	L.Push(kv)
	return 1
}

func Preload(ip Preloader) {
	kv := lua.NewUserKV()
	kv.Set("map", lua.NewFunction(NewMapL))
	kv.Set("slice", lua.NewFunction(NewSliceL))
	kv.Set("fmt", lua.NewFunction(NewFmtL))
	kv.Set("pretty", lua.NewFunction(PrettyJsonL))
	ip.Set("luakit", kv)
}
