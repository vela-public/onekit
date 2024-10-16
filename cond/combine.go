package cond

import "github.com/vela-public/onekit/lua"

type Combine []*Cond

func NewCombine() *Combine {
	return new(Combine)
}

func (iv *Combine) Add(cnd *Cond) {
	v := *iv
	v = append(v, cnd)
	*iv = v
}

func (iv *Combine) CheckMany(L *lua.LState, opt ...OptionFunc) {
	v := *iv
	cnd := CheckMany(L, opt...)
	v = append(v, cnd)
	*iv = v
}

func (iv *Combine) Match(data interface{}, opt ...OptionFunc) bool {
	v := *iv
	if len(v) == 0 {
		return true
	}

	for _, cnd := range v {
		if cnd.Match(data, opt...) {
			return true
		}
	}
	return false
}

func (iv *Combine) MatchOR(data interface{}, opt ...OptionFunc) bool {
	v := *iv
	if len(v) == 0 {
		return true
	}
	for _, cnd := range v {
		if cnd.Match(data, opt...) {
			return true
		}
	}
	return false
}
