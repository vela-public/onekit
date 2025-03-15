package cond

import (
	"bytes"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
)

func (cnd *Cond) append(s *Section) {
	cnd.data = append(cnd.data, s)
}

func (cnd *Cond) Mode(L *lua.LState, idx int) (CndMode, bool) {
	v := L.Get(idx)
	if v.Type() != lua.LTObject {
		return AND, false
	}

	cm, ok := v.(CndMode)
	if !ok {
		return AND, false
	}
	return cm, true
}

func (cnd *Cond) CheckMany(L *lua.LState, opt ...OptionFunc) {
	ov := &option{co: L, seek: 0}
	for _, fn := range opt {
		fn(ov)
	}

	top := L.GetTop()
	if top-ov.seek <= 0 {
		return
	}

	cm, ok := cnd.Mode(L, ov.seek+1)
	if ok {
		cnd.mode.put(cm)
		ov.seek++
		if top-ov.seek <= 0 {
			return
		}
	}

	for idx := ov.seek + 1; idx <= top; idx++ {
		val := L.Get(idx)
		var sec *Section
		switch val.Type() {
		case lua.LTFunction:
			sec = NewSectionLFunc(L, val.(*lua.LFunction))
		case lua.LTGoCond:
			sec = NewSectionGoFunc(L, func(v any, optionFunc ...OptionFunc) bool {
				return val.(lua.GoCond[any])(v)
			})
		case lua.LTGoFuncErr:
			sec = NewSectionGoFunc(L, func(v interface{}, optionFunc ...OptionFunc) bool {
				err := val.(lua.GoFuncErr)(v)
				return err == nil
			})
		case lua.LTGoFuncStr:
			sec = NewSectionGoFunc(L, func(v any, optionFunc ...OptionFunc) bool {
				str := val.(lua.GoFuncStr)(v)
				return str == ""
			})
		case lua.LTGoFuncInt:
			sec = NewSectionGoFunc(L, func(v any, optionFunc ...OptionFunc) bool {
				num := val.(lua.GoFuncInt)(v)
				return num == 0
			})
		default:
			sec = NewSectionText(L.IsString(idx))
			if !sec.Ok() {
				L.RaiseError("condition compile fail %v", sec.err)
			}
		}
		cnd.append(sec)
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
		value: v,
		logic: AND,
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

	if ov.field == nil && ov.compare == nil {
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
