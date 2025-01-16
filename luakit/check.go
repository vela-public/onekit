package luakit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/todo"
)

func Check[T any](L *lua.LState, val lua.LValue) (t T) {
	if val.Type() != lua.LTObject {
		L.RaiseError("must object type , got:%s", val.Type().String())
		return
	}

	if v, ok := val.(T); !ok {
		L.RaiseError("must %T got %T", t, v)
		return v
	}

	return
}

func TypeFor(L *lua.LState) string {
	v := L.Get(1)
	var s string
	switch v.Type() {
	case lua.LTNil:
		s = fmt.Sprintf("lua:nil go:nil value:nil")
	case lua.LTInt:
		s = fmt.Sprintf("lua:int go:int value:%d", v.(lua.LInt))
	case lua.LTBool:
		s = fmt.Sprintf("lua:bool go:bool value:%v", v)
	case lua.LTFunction:
		s = fmt.Sprintf("lua:function go:func(*state) int")
	case lua.LTTable:
		s = fmt.Sprintf("lua:table")
	case lua.LTString:
		s = fmt.Sprintf("lua:string go:string value:%v", v)
	case lua.LTNumber:
		s = fmt.Sprintf("lua:number go:float64 value:%v", v)
	case lua.LTChannel:
		s = fmt.Sprintf("lua:channel go:channel value:%v", v)
	case lua.LTKv:
		s = fmt.Sprintf("lua:kv go:slice")
	case lua.LTThread:
		s = fmt.Sprintf("lua:thread go:*lua.LState")
	case lua.LTObject:
		s = fmt.Sprintf("lua:object go:%T value:%v", v, v)
	case lua.LTGeneric:
		s = fmt.Sprintf("lua:generic go:%T value:%v", v.(lua.GenericType).UnwrapData(), v)
	default:
		s = fmt.Sprintf("lua:object go:%T value:%v", v, v)
	}
	return todo.IF(len(s) > 100, s[:100], s)
}
