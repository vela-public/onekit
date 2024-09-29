package pipe

import (
	"github.com/vela-public/onekit/lua"
)

func (px *Chains) CheckMany(L *lua.LState, opt ...func(*Chains)) {
	for _, fn := range opt {
		fn(px)
	}

	n := L.GetTop()
	if n-px.seek < 0 {
		return
	}

	offset := n - px.seek
	switch offset {
	case 0:
		return
	case 1:
		px.LValue(L.Get(px.seek + 1))

	default:
		for idx := px.seek + 1; idx <= n; idx++ {
			px.LValue(L.Get(idx))
		}
	}

	return

}

func (px *Chains) Check(L *lua.LState, idx int) {
	px.LValue(L.Get(idx))
}

func Lua(L *lua.LState, opt ...func(*Chains)) (px *Chains) {
	px = New(opt...)
	px.vm = px.xEnv.Clone(L)

	n := L.GetTop()
	if n-px.seek < 0 {
		return
	}

	offset := n - px.seek
	switch offset {
	case 0:
		return px
	case 1:
		px.LValue(L.Get(px.seek + 1))

	default:
		for idx := px.seek + 1; idx <= n; idx++ {
			px.LValue(L.Get(idx))
		}
	}

	return
}
