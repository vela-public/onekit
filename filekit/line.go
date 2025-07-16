package filekit

import (
	"encoding/json"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/jsonkit"
	"github.com/vela-public/onekit/lua"
)

type LineKV struct {
	Key string
	Val any
}

type Line struct {
	File string // file
	Size int    // The size of the file
	Text []byte // The contents of the file
	Json *jsonkit.FastJSON
}

func (line *Line) Type() lua.LValueType                   { return lua.LTObject }
func (line *Line) AssertFloat64() (float64, bool)         { return float64(line.Size), true }
func (line *Line) AssertString() (string, bool)           { return cast.B2S(line.Text), true }
func (line *Line) AssertFunction() (*lua.LFunction, bool) { return lua.NewFunction(line.InfoL), true }
func (line *Line) Hijack(*lua.CallFrameFSM) bool          { return false }
func (line *Line) String() string {
	return cast.B2S(line.Text)
}

func (line *Line) FastJSON() *jsonkit.FastJSON {
	if line.Json != nil {
		return line.Json
	}

	t := &jsonkit.FastJSON{}
	t.ParseText(cast.B2S(line.Text))
	line.Json = t
	return t
}

func (line *Line) Set(data ...LineKV) error {
	obj := make(map[string]any)
	err := json.Unmarshal(line.Text, &obj)
	if err != nil {
		return err
	}
	for _, kv := range data {
		if kv.Val == nil {
			delete(obj, kv.Key)
		} else {
			obj[kv.Key] = kv.Val
		}
	}

	text, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	line.Text = text
	t := &jsonkit.FastJSON{}
	t.ParseText(cast.B2S(text))
	line.Json = t
	return nil
}

func (line *Line) Int(path string) int {
	return line.Json.Int(path)
}

func (line *Line) InfoL(L *lua.LState) int {
	kv := lua.NewUserKV()
	kv.Set("file", lua.LString(line.File))
	kv.Set("size", lua.LInt(line.Size))
	kv.Set("text", lua.LString(cast.B2S(line.Text)))
	L.Push(kv)
	return 1
}

func (line *Line) Index(L *lua.LState, key string) lua.LValue {
	return line.FastJSON().Index(L, key)
}
func (line *Line) NewIndex(L *lua.LState, key string, val lua.LValue) {
	line.FastJSON().NewIndex(L, key, val)
}

func (line *Line) Meta(L *lua.LState, key lua.LValue) lua.LValue {
	return line.FastJSON().Meta(L, key)
}

func (line *Line) NewMeta(L *lua.LState, key lua.LValue, val lua.LValue) {
	line.FastJSON().NewMeta(L, key, val)
}
