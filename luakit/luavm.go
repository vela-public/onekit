package luakit

import (
	"github.com/vela-public/onekit/lua"
	"sync"
)

type Pool struct {
	bucket sync.Pool
}

func LuaVM(debug bool, callback ...func(*lua.LState)) *Pool {

	fn := func() interface{} {
		co := lua.NewState(lua.Options{
			CallStackSize:       lua.CallStackSize,
			RegistrySize:        lua.RegistrySize,
			IncludeGoStackTrace: debug,
		})

		for _, call := range callback {
			call(co)
		}
		return co
	}

	return &Pool{
		bucket: sync.Pool{New: fn},
	}
}

func (p *Pool) Get() *lua.LState {
	return p.bucket.Get().(*lua.LState)
}

func (p *Pool) Clone(co *lua.LState) *lua.LState {
	co2 := p.bucket.Get().(*lua.LState)
	co2.Copy(co)
	return co2
}

func (p *Pool) Put(co *lua.LState) {
	co.Keepalive()
	p.bucket.Put(co)
}
