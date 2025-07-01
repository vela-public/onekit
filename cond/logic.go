package cond

import (
	"github.com/vela-public/onekit/lua"
)

const (
	AND Logic = 1 << iota
	OR
	UNARY
	CODE
)

type Logic uint64

func (c Logic) Text() string {
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

func (c Logic) String() string                         { return c.Text() }
func (c Logic) Type() lua.LValueType                   { return lua.LTObject }
func (c Logic) AssertFloat64() (float64, bool)         { return float64(c), true }
func (c Logic) AssertString() (string, bool)           { return c.Text(), true }
func (c Logic) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (c Logic) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (c *Logic) has(v Logic) bool {
	a := *c
	return a&v == v
}

func (c *Logic) put(v Logic) {
	*c = *c | v
}

func (c *Logic) undo(v Logic) {
	a := *c
	*c = a & (^v)
}

func (c *Logic) set(v Logic) {
	*c = v
}
