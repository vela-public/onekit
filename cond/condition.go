package cond

import (
	"github.com/vela-public/onekit/lua"
)

type Cond struct {
	mode *CndMode
	data []*Section
}

func New() *Cond {
	cnd := &Cond{
		mode: new(CndMode),
	}
	cnd.mode.put(AND)
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
