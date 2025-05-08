package render

import (
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
)

func NewRenderTextL(L *lua.LState) int {
	var r *Render
	text := L.CheckString(1)
	reader := func() (string, bool) {
		return text, false
	}
	if top := L.GetTop(); top >= 2 {
		tag := lua.Check[lua.GoFunction[*Render]](L, L.Get(2))
		r = NewRender(reader, tag)
	} else {
		r = NewRender(reader)
	}

	r.PrepareText()
	L.Push(r)
	return 1
}

func NewRenderTagL(L *lua.LState) int {
	left := L.CheckString(1)
	right := L.CheckString(2)
	if left == "" || right == "" {
		L.RaiseError("render tag must be not empty")
		return 0
	}
	fn := Tag(left, right)
	L.Push(lua.GoFunction[*Render](fn))
	return 1
}

func NewRenderFileL(L *lua.LState) int {
	ft := &FileTemplate{}
	tab := L.CheckTable(1)
	err := luakit.TableTo(L, tab, ft)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}

	var r *Render
	if top := L.GetTop(); top >= 2 {
		tag := lua.Check[lua.GoFunction[*Render]](L, L.Get(2))
		r = NewRender(ft.Reader, tag)
	} else {
		r = NewRender(ft.Reader)
	}

	r.need = true
	L.Push(r)
	return 1
}

func Preload(p lua.Preloader) {
	kv := lua.NewUserKV()
	kv.Set("tag", lua.NewFunction(NewRenderTagL))
	kv.Set("file", lua.NewFunction(NewRenderFileL))
	p.Set("render", lua.NewExport("lua.render.export", lua.WithFunc(NewRenderTextL), lua.WithTable(kv)))
}
