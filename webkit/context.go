package webkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
)

type HandleFunc func(ctx *WebContext)

type WebContext struct {
	Request *RequestCtx
}

func (w *WebContext) String() string                         { return fmt.Sprintf("http.context %p", w) }
func (w *WebContext) Type() lua.LValueType                   { return lua.LTObject }
func (w *WebContext) AssertFloat64() (float64, bool)         { return 0, false }
func (w *WebContext) AssertString() (string, bool)           { return "", false }
func (w *WebContext) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (w *WebContext) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (w *WebContext) Bind(v any) error {
	data := w.Request.Request.Body()
	return json.Unmarshal(data, v)
}

func (w *WebContext) Args() *fasthttp.Args {
	return w.Request.QueryArgs()
}

func (w *WebContext) Str(name string) string {
	return string(w.Request.QueryArgs().Peek(name))
}

func (w *WebContext) Int(name string) int {
	return w.Request.QueryArgs().GetUintOrZero(name)
}

func (w *WebContext) Bool(name string) bool {
	return w.Request.QueryArgs().GetBool(name)
}

func (w *WebContext) Has(name string) bool {
	return w.Request.QueryArgs().Has(name)
}

func (w *WebContext) SayGo(code int, body string) {
	w.Request.Response.SetStatusCode(code)
	w.Request.Response.SetBodyString(body)
}

func (w *WebContext) SayL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	var buf bytes.Buffer
	for i := 1; i <= n; i++ {
		buf.WriteString(L.Get(i).String())
	}
	w.Request.SetBody(buf.Bytes())
	return 0
}

func (w *WebContext) ExitL(L *lua.LState) int {
	code := L.CheckInt(1)
	w.Request.Response.SetStatusCode(code)
	L.Terminated()
	return 0
}

func (w *WebContext) SayFormatL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	body := luakit.Format(L, 0)
	w.Request.Response.SetBodyRaw(lua.S2B(body))
	return 0
}

func (w *WebContext) SayRawL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	var buf bytes.Buffer
	for i := 1; i <= n; i++ {
		buf.Write(lua.S2B(L.CheckString(i)))
	}
	w.Request.Response.SetBodyRaw(buf.Bytes())
	return 0
}

func (w *WebContext) AppendL(L *lua.LState) int {
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	for i := 1; i <= n; i++ {
		dat := L.CheckString(i)
		if len(dat) == 0 {
			continue
		}
		w.Request.Response.AppendBody(cast.S2B(dat))
	}
	return 0
}

func (w *WebContext) JsonL(L *lua.LState) int {
	lv := L.CheckAny(1)
	chunk, err := json.Marshal(lv)
	if err != nil {
		w.Request.Error(err.Error(), 500)
		return 0
	}

	w.Request.SetBody(chunk)
	return 0
}

func (w *WebContext) fileL(L *lua.LState) int {
	path := L.CheckString(1)
	w.Request.SendFile(path)
	return 0
}

func (w *WebContext) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "say":
		return lua.NewFunction(w.SayL)
	case "say_raw":
		return lua.NewFunction(w.SayRawL)
	case "append":
		return lua.NewFunction(w.AppendL)
	case "file":
		return lua.NewFunction(w.fileL)
	case "json":
		return lua.NewFunction(w.JsonL)
	case "sayf":
		return lua.NewFunction(w.SayFormatL)
	case "exit":
		return lua.NewFunction(w.ExitL)
	}

	return K2V(w.Request, key)
}

func NewWebContext(ctx *RequestCtx) *WebContext {
	return &WebContext{Request: ctx}
}
