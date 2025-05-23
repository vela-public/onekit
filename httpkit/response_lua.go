package httpkit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipe"
	"os"
	"strings"
)

func (r *Response) Type() lua.LValueType                   { return lua.LTObject }
func (r *Response) AssertFloat64() (float64, bool)         { return 0, false }
func (r *Response) AssertString() (string, bool)           { return "", false }
func (r *Response) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (r *Response) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (r *Response) OnErrorHandler(e error) {
	if r.OnError == nil {
		return
	}
	r.OnError.Invoke(e.Error())
}

func (r *Response) catch(L *lua.LState) int {

	if r.Err != nil {
		r.OnErrorHandler(r.Err)
		return 0
	}

	if r.RawResponse == nil {
		r.OnErrorHandler(r.Err)
		return 0
	}

	n := L.GetTop()
	code := r.StatusCode()
	for i := 1; i <= n; i++ {
		if L.CheckInt(i) == code {
			return 0
		}
	}
	r.OnErrorHandler(fmt.Errorf("%s request not found valid status code , got: %d body: %s", r.Request.URL, code, r.Body()))
	return 0
}

func (r *Response) caseL(L *lua.LState) lua.LValue {
	if r.Swh == nil {
		r.Swh = pipe.NewSwitch()
	}

	return r.Swh.Index(L, "case")
}

func (r *Response) saveL(L *lua.LState) int {
	path := L.CheckString(1)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		L.Push(lua.S2L(err.Error()))
		return 1
	}

	defer file.Close()

	err = r.RawResponse.Write(file)
	if err != nil {
		L.Push(lua.S2L(err.Error()))
		return 1
	}
	return 0
}

func (r *Response) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "body":
		return lua.B2L(r.Body())
	case "size":
		return lua.LInt(r.size)
	case "code":
		return lua.LInt(r.StatusCode())
	case "url":
		return lua.S2L(r.Request.URL)
	case "save":
		return lua.NewFunction(r.saveL)

	case "case":
		return r.caseL(L)

	case "catch":
		return L.NewFunction(r.catch)

	}

	if strings.HasPrefix(key, "http_") {
		return lua.S2L(r.Header().Get(U2H(key[5:])))
	}

	return lua.LNil
}
