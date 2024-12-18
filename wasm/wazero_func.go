package wasm

import (
	"fmt"
	"github.com/tetratelabs/wazero/api"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/jsonkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/pipe"
)

const (
	Number ResultType = iota
	Text
	Json
	Bin
)

type ResultType uint8

type Function struct {
	name   string
	parent *Module
	typ    ResultType
	mask   []BinaryMask
	api    api.Function
	handle *pipe.Chain
}

func (fn *Function) Write(buff []byte) (int, error) {
	if fn.api == nil {
		return 0, fmt.Errorf("not found wasm func")
	}

	p, s := StringToPtr(cast.B2S(buff))
	arg := uint64(p)<<32 | uint64(s)

	ret, err := fn.api.Call(fn.parent.ctx, arg)
	if err != nil {
		return 0, err
	}

	page := ret[0]
	var data any
	switch fn.typ {
	case Number: //number
		data = lua.LNumber(page)
	case Text: //text
		chunk, e := fn.reader(page)
		if e != nil {
			return 0, e
		}
		data = chunk

	case Json: //Json
		f := &jsonkit.FastJSON{}
		chunk, e := fn.reader(page)
		if e != nil {
			fn.parent.NoError(e)
			return 0, e
		}
		f.ParseText(cast.B2S(chunk))
		data = f

	case Bin: //Bin
		chunk, _ := fn.reader(page)
		data = &BinaryType{fn.mask, chunk}
	default:
		return 0, fmt.Errorf("not found data type")
	}

	ctx := fn.handle.Invoke(data)
	if e := ctx.UnwrapErr(); e != nil {
		return 0, e
	}

	return len(buff), nil
}

func (fn *Function) lock() {
	fn.parent.mutex.Lock()
}
func (fn *Function) unlock() {
	fn.parent.mutex.Unlock()
}

func (fn *Function) String() string {
	return fmt.Sprintf("mod.%s.function", fn.name)
}

func (fn *Function) Type() lua.LValueType              { return lua.LTObject }
func (fn *Function) AssertFloat64() (float64, bool)    { return 0, false }
func (fn *Function) AssertString() (string, bool)      { return "", false }
func (fn *Function) Hijack(fsm *lua.CallFrameFSM) bool { return false }
func (fn *Function) AssertFunction() (*lua.LFunction, bool) {
	return lua.NewFunction(fn.Call), true
}

func (fn *Function) pipeL(L *lua.LState) int {
	sub := pipe.Lua(L, pipe.LState(L), pipe.Reuse(L, true))
	fn.handle.Merge(sub)
	L.Push(fn)
	return 1
}

func (fn *Function) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "json":
		return lua.NewFunction(func(L *lua.LState) int {
			fn.typ = Json
			L.Push(fn)
			return 1
		})
	case "pipe":
		return lua.NewFunction(fn.pipeL)

	case "text":
		return lua.NewFunction(func(L *lua.LState) int {
			fn.typ = Text
			L.Push(fn)
			return 1
		})
	case "number":
		return lua.NewFunction(func(L *lua.LState) int {
			fn.typ = Number
			L.Push(fn)
			return 1
		})

	case "binary":
		return lua.NewFunction(func(L *lua.LState) int {
			fn.typ = Bin
			tab := L.CheckTable(1)
			fn.mask = make([]BinaryMask, 0)
			err := luakit.TableTo(L, tab, &fn.mask)
			if err != nil {
				L.RaiseError("mask error:%s", err.Error())
			}

			L.Push(fn)
			return 1
		})

	}
	return lua.LNil
}

func (fn *Function) reader(page uint64) ([]byte, error) {
	ptr := uint32(page >> 32)
	sz := uint32(page)
	buff, ok := fn.parent.mod.Memory().Read(ptr, sz)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range", ptr, sz)
	}
	return buff, nil
}

func (fn *Function) ParamL(L *lua.LState) []uint64 {
	var params []uint64
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		lv := L.Get(i)
		switch lv.Type() {
		case lua.LTNumber:
			params = append(params, uint64(lv.(lua.LNumber)))
		case lua.LTInt:
			params = append(params, uint64(lv.(lua.LInt)))
		case lua.LTUint:
			params = append(params, uint64(lv.(lua.LUint)))
		case lua.LTUint64:
			params = append(params, uint64(lv.(lua.LUint64)))
		case lua.LTInt64:
			params = append(params, uint64(lv.(lua.LInt64)))
		case lua.LTString, lua.LTObject:
			text := lv.String()
			ptr, sz, err := fn.parent.Buffer(text)
			if err != nil {
				L.RaiseError("Memory overflow max:%d", fn.parent.mod.Memory().Size())
				return nil
			}

			params = append(params, uint64(ptr)<<32|uint64(sz))
			//params = append(params, uint64(ptr), uint64(sz))
		default:
			L.RaiseError("%v type not support", lv.Type())
			return nil
		}

	}
	return params
}

func (fn *Function) Call(L *lua.LState) int {
	if fn.api == nil {
		L.RaiseError("Function %s not found", fn.name)
	}

	fn.lock()
	defer fn.unlock()

	params := fn.ParamL(L)
	ret, err := fn.api.Call(fn.parent.ctx, params...)
	if err != nil {
		fn.parent.NoError(err)
		return 0
	}

	page := ret[0]
	switch fn.typ {
	case Number: //number
		L.Push(lua.LNumber(page))
		return 1
	case Text: //text
		buff, e := fn.reader(page)
		if e != nil {
			fn.parent.NoError(e)
			return 0
		}
		L.Push(lua.B2L(buff))
		return 1

	case Json: //Json
		f := &jsonkit.FastJSON{}
		buff, e := fn.reader(page)
		if e != nil {
			fn.parent.NoError(e)
			return 0
		}
		f.ParseText(cast.B2S(buff))
		L.Push(f)
		return 1

	case Bin: //Bin
		buff, _ := fn.reader(page)
		L.Push(&BinaryType{fn.mask, buff})
		return 1
	default:
		return 0
	}
}
