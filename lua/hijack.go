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
		v, data := Unwrap(fsm.co.reg.Get(base + int(fsm.inst&0x1ff)))
		switch dat := data.(type) {
		case IndexType:
			return fsm.Index(dat.Index)
		case Getter:
			return fsm.Index(func(_ *LState, key string) LValue {
				return ReflectTo(dat.Getter(key))
			})
		default:
			if da, ok := v.(IndexOfType); ok {
				return fsm.Index(da.IndexOf)
			}
			return v.Hijack(fsm)
		}

	case OP_GETTABLEKS:
		base := fsm.co.currentFrame.LocalBase
		v, data := Unwrap(fsm.co.reg.Get(base + int(fsm.inst&0x1ff)))
		switch dat := data.(type) {
		case IndexType:
			return fsm.Index(dat.Index)
		case Getter:
			return fsm.Index(func(_ *LState, key string) LValue {
				return ReflectTo(dat.Getter(key))
			})
		default:
			if da, ok := v.(IndexOfType); ok {
				return fsm.Index(da.IndexOf)
			}
			return v.Hijack(fsm)
		}

	case OP_GETTABLE:
		base := fsm.co.currentFrame.LocalBase
		v, data := Unwrap(fsm.co.Get(base + int(fsm.inst&0x1ff)))
		switch dat := data.(type) {
		case MetaType:
			return fsm.Meta(dat.Meta)
		case Getter:
			return fsm.Meta(func(_ *LState, k LValue) LValue {
				return ReflectTo(k.String())
			})
		default:
			if da, ok := v.(MetaOfType); ok {
				return fsm.Meta(da.MetaOf)
			}
			return v.Hijack(fsm)
		}

	case OP_SETTABLEKS:
		base := fsm.co.currentFrame.LocalBase
		v, data := Unwrap(fsm.co.reg.Get(base + (int(fsm.inst>>18) & 0xff)))
		switch dat := data.(type) {
		case NewIndexType:
			return fsm.NewIndex(dat.NewIndex)
		case Setter:
			return fsm.NewIndex(func(_ *LState, k string, v LValue) {
				dat.Setter(k, v)
			})
		default:
			if da, ok := v.(NewIndexOfType); ok {
				return fsm.NewIndex(da.NewIndexOf)
			}
			return v.Hijack(fsm)
		}

	case OP_SETTABLE:
		base := fsm.co.currentFrame.LocalBase
		v, data := Unwrap(fsm.co.reg.Get(base + int(fsm.inst>>18)&0xff))
		switch dat := data.(type) {
		case NewMetaType:
			return fsm.NewMeta(dat.NewMeta)
		case Setter:
			return fsm.NewMeta(func(_ *LState, k LValue, v LValue) {
				dat.Setter(k.String(), v)
			})
		default:
			if da, ok := v.(NewMetaOfType); ok {
				return fsm.NewMeta(da.NewMetaOf)
			}
			return v.Hijack(fsm)
		}

	default:
		return false
	}

}
