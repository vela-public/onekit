package cond

import (
	"github.com/vela-public/onekit/lua"
)

type Cond struct {
	logic *Logic
	data  []*Section
}

func New() *Cond {
	cnd := &Cond{
		logic: new(Logic),
	}
	cnd.logic.put(AND)
	return cnd
}

func (cnd *Cond) Text(v ...string) {
	sz := len(v)
	if sz == 0 {
		return
	}

	cnd.data = make([]*Section, sz)

	for i := 0; i < sz; i++ {
		cnd.data[i] = NewSectionText(v[i])
	}
}

func NewText(v ...string) *Cond {
	cnd := New()
	cnd.Text(v...)
	return cnd
}

func CheckMany(L *lua.LState, opt ...OptionFunc) *Cond {
	cnd := &Cond{}
	cnd.CheckMany(L, opt...)
	return cnd
}
