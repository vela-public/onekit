package pipe

import (
	"github.com/vela-public/onekit/lua"
)

type Environment interface {
	Coroutine() *lua.LState
	Clone(*lua.LState) *lua.LState
	Free(*lua.LState)
	Errorf(string, ...interface{})
}

type Handler func(...interface{}) error

type Chains struct {
	chain []Handler
	seek  int
	vm    *lua.LState
	xEnv  Environment
}

func (px *Chains) Merge(sub *Chains) {
	if sub.Len() == 0 {
		return
	}
	px.chain = append(px.chain, sub.chain...)
}

func (px *Chains) clone(co *lua.LState) *lua.LState {
	return px.xEnv.Clone(co)
}
func (px *Chains) append(v Handler) {
	if v == nil {
		return
	}

	px.chain = append(px.chain, v)
}

func (px *Chains) coroutine() *lua.LState {
	return px.xEnv.Coroutine()
}

func (px *Chains) free(co *lua.LState) {
	px.xEnv.Free(co)
}

func (px *Chains) invalid(format string, v ...interface{}) {
	if px.xEnv == nil {
		//vela.GxEnv().Errorf(format, v...)
		return
	}

	px.xEnv.Errorf(format, v...)
}

func New(opt ...func(*Chains)) (px *Chains) {
	px = &Chains{}

	for _, fn := range opt {
		fn(px)
	}

	return
}
