package httpkit

import (
	"bytes"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
	"strings"
)

func (r *RawResponse) String() string                         { return cast.B2S(r.Bytes()) }
func (r *RawResponse) Type() lua.LValueType                   { return lua.LTObject }
func (r *RawResponse) AssertFloat64() (float64, bool)         { return 0, false }
func (r *RawResponse) AssertString() (string, bool)           { return r.String(), true }
func (r *RawResponse) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (r *RawResponse) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (r *RawResponse) Bytes() []byte {
	if r == nil {
		return nil
	}

	var buf bytes.Buffer

	buf.WriteString(r.rawStatus)
	buf.WriteString("\r\n")
	for _, v := range r.headers {
		buf.WriteString(v)
		buf.WriteString("\r\n")
	}
	buf.WriteString("\r\n")

	buf.Write(r.body)
	return buf.Bytes()
}

func (r *RawResponse) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "code":
		code := r.StatusCode()

		return lua.LInt(cast.ToInt(code))
	case "body":
		return lua.B2L(r.body)

	}

	if strings.HasPrefix(key, "h_") {
		return lua.S2L(r.Header(key[3:]))
	}

	return lua.LNil
}
