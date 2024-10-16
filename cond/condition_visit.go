package cond

import (
	"bytes"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
)

func (cnd *Cond) append(s *Section) {
	cnd.data = append(cnd.data, s)
}

func (cnd *Cond) CheckMany(L *lua.LState, opt ...OptionFunc) {
	ov := &option{}
	for _, fn := range opt {
		fn(ov)
	}

	n := L.GetTop()
	offset := n - ov.seek
	if offset < 0 {
		return
	}

	if ov.unary {
		sec := &Section{method: Unary}
		for idx := ov.seek + 1; idx <= n; idx++ {
			lv := L.Get(idx)
			switch lv.Type() {
			case lua.LTFunction:
				sec.method = Fn
			case lua.LTGoFuncErr:
			case lua.LTGoFuncStr:
			case lua.LTGoFuncInt:

			default:
				if item := lua.IsString(lv); len(item) > 0 {
					sec.data = append(sec.data, item)
				}
			}
		}

		if len(sec.data) == 0 {
			L.RaiseError("condition compile fail not found data")
			return
		}

		cnd.append(sec)
		return
	}

	switch offset {
	case 0:
		return
	case 1:
		sec := Compile(L.IsString(ov.seek + 1))
		if sec.Ok() {
			cnd.append(sec)
			return
		}
		L.RaiseError("condition compile fail %v", sec.err)

	default:
		for idx := ov.seek + 1; idx <= n; idx++ {
			sec := Compile(L.IsString(idx))
			if sec.Ok() {
				cnd.append(sec)
				continue
			}
			L.RaiseError("condition compile fail %v", sec.err)
		}
	}

	return
}

func (cnd *Cond) Len() int {
	return len(cnd.data)
}

func (cnd *Cond) String() string {
	if cnd.Len() == 0 {
		return ""
	}

	n := cnd.Len()
	var buf bytes.Buffer
	for i := 0; i < n; i++ {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(cnd.data[i].raw)
	}
	return buf.String()
}

func (cnd *Cond) matchOr(ov *option, n int) bool {
	for i := 0; i < n; i++ {
		sec := cnd.data[i]
		ok, err := sec.Call(ov)
		if err != nil {
			ov.errs.Try(sec.raw, err)
			continue
		}

		if ok {
			return true
		}
	}

	return false
}

func (cnd *Cond) matchAnd(ov *option, n int) bool {
	flag := false
	for i := 0; i < n; i++ {
		sec := cnd.data[i]
		ok, err := sec.Call(ov)
		if err != nil {
			ov.errs.Try(sec.raw, err)
			continue
		}

		if !ok {
			return false
		} else {
			flag = true
		}
	}

	return flag
}

func (cnd *Cond) with(v interface{}, opt ...OptionFunc) *option {
	ov := &option{
		logic: AND,
		value: v,
		errs:  errkit.Errors(),
	}
	for _, fn := range opt {
		fn(ov)
	}
	ov.NewPeek(v)
	return ov
}

func (cnd *Cond) Match(v interface{}, opt ...OptionFunc) bool {
	n := cnd.Len()
	if n == 0 {
		return true
	}

	ov := cnd.with(v, opt...)

	if ov.peek == nil && ov.compare == nil {
		return false
	}

	switch ov.logic {
	case AND:
		return cnd.matchAnd(ov, n)
	case OR:
		return cnd.matchOr(ov, n)

	default:
		return false

	}
}

func (cnd *Cond) Merge(v *Cond) {
	if len(v.data) == 0 {
		return
	}
	cnd.data = append(cnd.data, v.data...)
}

func (cnd *Cond) Append(v ...*Section) {
	cnd.data = append(cnd.data, v...)
}
