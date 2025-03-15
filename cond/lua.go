package cond

import (
	"github.com/vela-public/onekit/lua"
)

func Preload(v lua.Preloader) {
	tab := lua.NewUserKV()
	tab.Set("AND", AND)
	tab.Set("OR", OR)
	tab.Set("UNARY", UNARY)
	tab.Set("CODE", CODE)
	v.Set("cnd", lua.NewExport("lua.cnd.export", lua.WithFunc(NewCondL)))
}
