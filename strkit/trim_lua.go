package strkit

import "github.com/vela-public/onekit/lua"

func (t *Trim) String() string                         { return t.Text }
func (t *Trim) Type() lua.LValueType                   { return lua.LTObject }
func (t *Trim) AssertFloat64() (float64, bool)         { return 0, false }
func (t *Trim) AssertString() (string, bool)           { return "", false }
func (t *Trim) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (t *Trim) Hijack(*lua.CallFrameFSM) bool          { return false }

func (t *Trim) push(L *lua.LState) int {
	L.Push(t)
	return 1
}

func (t *Trim) fileL(L *lua.LState) int {
	filename := L.CheckFile(1)
	sub := L.CheckString(2)

	t.Handle = append(t.Handle, NewTrimAcFile(filename, sub))
	return t.push(L)
}

func (t *Trim) dateL(L *lua.LState) int {
	sub := L.CheckString(1)
	fm := L.Get(2)

	t.Handle = append(t.Handle, NewTrimDate(sub, fm.String()))
	return t.push(L)
}

func (t *Trim) numL(L *lua.LState) int {
	t.Handle = append(t.Handle, NewTrimN())
	return t.push(L)
}

func (t *Trim) graphicL(L *lua.LState) int {
	flag := L.IsTrue(1)
	t.Handle = append(t.Handle, NewTrimGraphic(flag))
	return t.push(L)
}

func (t *Trim) spaceL(L *lua.LState) int {
	t.Handle = append(t.Handle, NewTrimSpace())
	return t.push(L)
}

func (t *Trim) regexL(L *lua.LState) int {
	regex := L.CheckString(1)
	sub := L.CheckString(2)
	t.Handle = append(t.Handle, NewTrimRegex(regex, sub))
	return t.push(L)
}

func (t *Trim) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "mark":
		return lua.S2L(t.ToMask())

	case "date":
		return L.NewFunction(t.dateL)
	case "num":
		return L.NewFunction(t.numL)
	case "graphic":
		return L.NewFunction(t.graphicL)
	case "space":
		return L.NewFunction(t.spaceL)
	case "regex":
		return L.NewFunction(t.regexL)
	case "file":
		return L.NewFunction(t.fileL)
	}

	return lua.LNil
}

func NewTrimL(L *lua.LState) int {
	str := L.CheckString(1)
	gen := &Trim{Text: str}
	L.Push(gen)
	return 1
}
