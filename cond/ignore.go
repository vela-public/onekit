package cond

import "github.com/vela-public/onekit/lua"

type Ignore []*Cond

func NewIgnore() *Ignore {
	return new(Ignore)
}

func (iv *Ignore) Add(cnd *Cond) {
	v := *iv
	v = append(v, cnd)
	*iv = v
}

func (iv *Ignore) CheckMany(L *lua.LState, opt ...OptionFunc) {
	v := *iv
	cnd := CheckMany(L, opt...)
	v = append(v, cnd)
	*iv = v
}

func (iv *Ignore) Match(data interface{}, opt ...OptionFunc) bool {
	v := *iv
	if len(v) == 0 {
		return false
	}

	for _, cnd := range v {
		if cnd.Match(data, opt...) {
			return true
		}
	}
	return false
}
