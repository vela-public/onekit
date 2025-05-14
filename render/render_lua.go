package render

import "github.com/vela-public/onekit/lua"

func (r *Render) String() string                 { text, _ := r.Reader(); return text }
func (r *Render) Type() lua.LValueType           { return lua.LTObject }
func (r *Render) AssertFloat64() (float64, bool) { return 0, false }
func (r *Render) AssertString() (string, bool)   { return "", false }
func (r *Render) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(r.execL), true
}

func (r *Render) Hijack(fsm *lua.CallFrameFSM) bool { return false }

func (r *Render) execL(L *lua.LState) int {
	text := r.Render(L.Get(1), &Env{LState: L})
	L.Push(lua.S2L(text))
	return 1
}

func (r *Render) Index(L *lua.LState, key string) lua.LValue {
	return lua.LNil
}

func (r *Render) NewIndex(L *lua.LState, key string, val lua.LValue) {
	r.DataKV.Set(key, val)
}
