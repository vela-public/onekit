package cond

import (
	"github.com/vela-public/onekit/lua"
)

type LCond struct {
	//happy *pipe.Chain
	//sorry *pipe.Chain
	cnd *Cond
}

func (lc *LCond) String() string                         { return "condition" }
func (lc *LCond) Type() lua.LValueType                   { return lua.LTObject }
func (lc *LCond) AssertFloat64() (float64, bool)         { return 0, false }
func (lc *LCond) AssertString() (string, bool)           { return "", false }
func (lc *LCond) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (lc *LCond) Hijack(*lua.CallFrameFSM) bool          { return false }

func (lc *LCond) Match(lv lua.LValue, L *lua.LState) bool {
	if lc.cnd.Match(lv) {
		return true
	}

	return false
}

func (lc *LCond) matchL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		L.Push(lua.LFalse)
		return 1
	}

	for i := 1; i <= n; i++ {
		if lc.Match(L.Get(i), L) {
			L.Push(lua.LTrue)
			return 1
		}
	}

	L.Push(lua.LFalse)
	return 1
}

func (lc *LCond) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "match":
		return lua.NewFunction(lc.matchL)
	}

	return lua.LNil
}

func NewL(L *lua.LState) *LCond {
	return &LCond{
		cnd: CheckMany(L, Seek(0)),
	}
}

func NewCondL(L *lua.LState) int {
	L.Push(NewL(L))
	return 1
}

func Preload(v lua.Preloader) {
	v.Set("cnd", lua.NewExport("lua.cnd.export", lua.WithFunc(NewCondL)))

}
