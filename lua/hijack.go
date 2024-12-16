package lua

type CallFrameFSM struct {
	co   *LState
	op   int
	inst uint32
	base *callFrame
}

func (fsm *CallFrameFSM) LVAsBool(flag bool) {
	C := int(fsm.inst>>9) & 0x1ff //GETC
	if (C == 0) == flag {
		fsm.co.currentFrame.Pc++
	}
}

func (fsm *CallFrameFSM) Push(lv LValue) {
	reg := fsm.co.reg
	base := fsm.co.currentFrame.LocalBase
	RA := base + (int(fsm.inst>>18) & 0xff)
	reg.Set(RA, lv)
}

func (fsm *CallFrameFSM) OpCode() int {
	return fsm.op
}

func (fsm *CallFrameFSM) Index(hook func(*LState, string) LValue) bool {
	switch fsm.op {
	case OP_GETTABLEKS, OP_SELF:
	default:
		return false
	}

	reg := fsm.co.reg
	base := fsm.co.currentFrame.LocalBase
	RA := base + (int(fsm.inst>>18) & 0xff)
	C := int(fsm.inst>>9) & 0x1ff //GETC
	reg.Set(RA, hook(fsm.co, fsm.co.rkString(C)))
	return true
}

func (fsm *CallFrameFSM) NewIndex(hook func(*LState, string, LValue)) bool {
	if fsm.op != OP_SETTABLEKS {
		return false
	}

	B := int(fsm.inst & 0x1ff)    //GETB
	C := int(fsm.inst>>9) & 0x1ff //GETC
	key := fsm.co.rkString(B)
	val := fsm.co.rkValue(C)

	hook(fsm.co, key, val)
	return true
}

func (fsm *CallFrameFSM) NewMeta(hook func(*LState, LValue, LValue)) bool {
	if fsm.op != OP_SETTABLEKS {
		return false
	}
	B := int(fsm.inst & 0x1ff)    //GETB
	C := int(fsm.inst>>9) & 0x1ff //GETC
	key := fsm.co.rkValue(B)
	val := fsm.co.rkValue(C)
	hook(fsm.co, key, val)
	return true
}

func (fsm *CallFrameFSM) Meta(hook func(*LState, LValue) LValue) bool {
	if fsm.op != OP_GETTABLE {
		return false
	}

	reg := fsm.co.reg
	RA := fsm.co.currentFrame.LocalBase + (int(fsm.inst>>18) & 0xff)
	C := int(fsm.inst>>9) & 0x1ff //GETC
	reg.Set(RA, hook(fsm.co, fsm.co.rkValue(C)))
	return true
}

func HijackTable(fsm *CallFrameFSM) bool {
	switch fsm.op {
	case OP_SELF:
		base := fsm.co.currentFrame.LocalBase
		data := fsm.co.reg.Get(base + int(fsm.inst&0x1ff))
		if v, ok := data.(IndexEx); ok {
			return fsm.Index(v.Index)
		}

		if v, ok := data.(Getter); ok {
			return fsm.Index(func(_ *LState, key string) LValue {
				ret := v.Getter(key)
				return ToLValue(ret)
			})
		}

		return data.Hijack(fsm)

	case OP_GETTABLEKS:
		base := fsm.co.currentFrame.LocalBase
		data := fsm.co.reg.Get(base + int(fsm.inst&0x1ff))

		if v, ok := data.(IndexEx); ok {
			return fsm.Index(v.Index)
		}

		if v, ok := data.(Getter); ok {
			return fsm.Index(func(_ *LState, key string) LValue {
				ret := v.Getter(key)
				return ToLValue(ret)
			})
		}

		return data.Hijack(fsm)

	case OP_GETTABLE:
		base := fsm.co.currentFrame.LocalBase
		data := fsm.co.Get(base + int(fsm.inst&0x1ff))
		if v, ok := data.(MetaEx); ok {
			return fsm.Meta(v.Meta)
		}

		if v, ok := data.(Getter); ok {
			return fsm.Meta(func(_ *LState, value LValue) LValue {
				ret := v.Getter(value.String())
				return ToLValue(ret)

			})
		}

		return data.Hijack(fsm)

	case OP_SETTABLEKS:
		base := fsm.co.currentFrame.LocalBase
		data := fsm.co.reg.Get(base + (int(fsm.inst>>18) & 0xff))
		if v, ok := data.(NewIndexEx); ok {
			return fsm.NewIndex(v.NewIndex)
		}

		if v, ok := data.(Setter); ok {
			return fsm.NewIndex(func(_ *LState, key string, val LValue) {
				v.Setter(key, val)
			})
		}
		return data.Hijack(fsm)

	case OP_SETTABLE:
		base := fsm.co.currentFrame.LocalBase
		data := fsm.co.reg.Get(base + int(fsm.inst>>18)&0xff)
		if v, ok := data.(NewMetaEx); ok {
			return fsm.NewMeta(v.NewMeta)
		}

		if v, ok := data.(Setter); ok {
			return fsm.NewMeta(func(_ *LState, key LValue, val LValue) {
				v.Setter(key.String(), val)
			})
		}

		return data.Hijack(fsm)

	default:
		return false
	}

}
