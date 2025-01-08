package wasm

import "github.com/vela-public/onekit/lua"

func wasmL(L *lua.LState) int {
	file := L.CheckString(1)
	ctx := L.Context()

	m, e := NewModule(ctx, file)
	if e != nil {
		L.RaiseError("%v", e)
		return 0
	}
	L.Push(m)
	return 1
}

func Preload(v lua.Preloader) { //luakit
	v.Set("wasm", lua.NewExport("lua.wasm.export", lua.WithFunc(wasmL)))
}
