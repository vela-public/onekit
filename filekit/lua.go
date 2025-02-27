package filekit

import "github.com/vela-public/onekit/lua"

func NewFileKitL(L *lua.LState) int {
	return 0
}

func IndexL(L *lua.LState, key string) lua.LValue {
	return lua.LNil
}

func Preload(p lua.Preloader) {
	p.SetGlobal("filekit", lua.NewExport("lua.filekit.export", lua.WithFunc(NewFileKitL), lua.WithIndex(IndexL)))
}
