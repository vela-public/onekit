package cond

import (
	"github.com/vela-public/onekit/lua"
)

const (
	AND CndMode = 1 << iota
	OR
	UNARY
	CODE
)

type CndMode uint64

func (c CndMode) Text() string {
	switch c {
	case AND:
		return "and"
	case OR:
		return "or"
	case UNARY:
		return "unary"
	case CODE:
		return "code"
	default:
		return "unknown"
	}
}

func (c CndMode) String() string                         { return c.Text() }
func (c CndMode) Type() lua.LValueType                   { return lua.LTObject }
func (c CndMode) AssertFloat64() (float64, bool)         { return float64(c), true }
func (c CndMode) AssertString() (string, bool)           { return c.Text(), true }
func (c CndMode) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c CndMode) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (c *CndMode) has(v CndMode) bool {
	a := *c
	return a&v == v
}

func (c *CndMode) put(v CndMode) {
	*c = *c | v
}

func (c *CndMode) undo(v CndMode) {
	a := *c
	*c = a & (^v)
}

func (c *CndMode) set(v CndMode) {
	*c = v
}
