package taskit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
)

func (srv *Service) String() string                         { return fmt.Sprintf("%p", srv) }
func (srv *Service) Type() lua.LValueType                   { return lua.LTService }
func (srv *Service) AssertFloat64() (float64, bool)         { return 0, false }
func (srv *Service) AssertString() (string, bool)           { return "", false }
func (srv *Service) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (srv *Service) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (srv *Service) Text() string {
	if srv.data == nil {
		return fmt.Sprintf("task.service<%p>", srv)
	}

	str, ok := srv.data.(fmt.Stringer)
	if ok {
		return str.String()
	}

	return fmt.Sprintf("task.service<%p>", srv)
}

func (srv *Service) Private(L *lua.LState) {
	tas := CheckTaskEx(L, func(err error) {
		L.RaiseError("%v", err)
	})

	if tas.Key() != srv.from {
		L.RaiseError("current.task=%s service.from=%s with %s not allow", tas.Key(), srv.from, srv.name)
		return
	}
	srv.private = true
}

func (srv *Service) Index(L *lua.LState, key string) lua.LValue {
	if srv.data == nil {
		return lua.LNil
	}

	it, ok := srv.data.(lua.IndexType)
	if ok {
		return it.Index(L, key)
	}
	return lua.LNil
}

func (srv *Service) Meta(L *lua.LState, key lua.LValue) lua.LValue {
	if srv.data == nil {
		return lua.LNil
	}

	it, ok := srv.data.(lua.MetaType)
	if ok {
		return it.Meta(L, key)
	}
	return lua.LNil
}

func (srv *Service) NewMeta(L *lua.LState, key lua.LValue, val lua.LValue) {
	if srv.data == nil {
		return
	}

	it, ok := srv.data.(lua.NewMetaType)
	if ok {
		it.NewMeta(L, key, val)
		return
	}
}

func (srv *Service) MetaTable(L *lua.LState, key string) lua.LValue {
	if srv.data == nil {
		return lua.LNil
	}
	it, ok := srv.data.(lua.MetaTableType)
	if ok {
		return it.MetaTable(L, key)
	}
	return lua.LNil
}

func (srv *Service) NewIndex(L *lua.LState, key string, val lua.LValue) {
	if srv.data == nil {
		return
	}

	it, ok := srv.data.(lua.NewIndexType)
	if ok {
		it.NewIndex(L, key, val)
		return
	}
}
