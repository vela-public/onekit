package cond

import (
	"bytes"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
)

const (
	OR Logic = iota + 1
	AND
)

type Logic uint8

func (l Logic) String() string {
	switch l {
	case OR:
		return "or"
	case AND:
		return "and"
	}
	return "unknown"
}

type Cond struct {
	data []*Section
}

func New(c ...string) *Cond {
	n := len(c)
	if n == 0 {
		return &Cond{}
	}

	cond := &Cond{
		data: make([]*Section, len(c)),
	}

	for i := 0; i < n; i++ {
		cond.data[i] = Compile(c[i])
	}
	return cond
}

func F(prefix string, v ...interface{}) string {
	if len(v) == 0 {
		return ""
	}

	var buf bytes.Buffer
	buf.Write(cast.S2B(prefix))
	buf.WriteByte(' ')
	for i, item := range v {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(cast.ToString(item))
	}

	return buf.String()

}

func CheckMany(L *lua.LState, opt ...OptionFunc) *Cond {
	cnd := &Cond{}
	cnd.CheckMany(L, opt...)
	return cnd
}
