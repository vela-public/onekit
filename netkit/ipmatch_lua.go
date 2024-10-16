package netkit

import "github.com/vela-public/onekit/lua"

func (ipm *IPMatch) String() string                 { return "ip.matcher" }
func (ipm *IPMatch) Type() lua.LValueType           { return lua.LTObject }
func (ipm *IPMatch) AssertFloat64() (float64, bool) { return 0, false }
func (ipm *IPMatch) AssertString() (string, bool)   { return "", false }
func (ipm *IPMatch) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(ipm.MatchL), true
}
func (ipm *IPMatch) Hijack(fsm *lua.CallFrameFSM) bool { return false }
func (ipm *IPMatch) FromFileL(L *lua.LState) int {
	path := L.CheckFile(1)

	err := ipm.File(path)
	if err != nil {
		if top := L.GetTop(); top == 2 && L.IsTrue(2) { //must
			return 0
		} else {
			L.RaiseError(err.Error())
		}
	}
	L.Push(ipm)
	return 1
}

func (ipm *IPMatch) MatchL(L *lua.LState) int {
	ip := L.CheckString(1)
	L.Push(lua.LBool(ipm.Match(ip)))
	return 1
}

func (ipm *IPMatch) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "file":
		return lua.NewFunction(ipm.FromFileL)
	case "match":
		return lua.NewFunction(ipm.MatchL)
	}
	return lua.LNil
}

func NewIPMatchL(L *lua.LState) int {
	L.Push(&IPMatch{})
	return 1
}
