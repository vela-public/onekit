package luakit

import (
	"context"
	"github.com/vela-public/onekit/lua"
	"sync"
)

type Pool struct {
	bucket sync.Pool
}

func Thread(L *lua.LState) *Pool {
	fn := func() any {
		co, _ := L.NewThread()
		return co
	}

	return &Pool{
		bucket: sync.Pool{New: fn},
	}
}

func LuaVM(ctx context.Context, debug bool, callback func(*lua.LState), options ...func(lua.Preloader)) *Pool {
	kit := Apply(options...)
	fn := func() interface{} {
		co := kit.NewState(ctx, lua.Options{
			CallStackSize:       lua.CallStackSize,
			RegistrySize:        lua.RegistrySize,
			IncludeGoStackTrace: debug,
		})

		callback(co)
		return co
	}

	return &Pool{
		bucket: sync.Pool{New: fn},
	}
}

func (p *Pool) Get() *lua.LState {
	return p.bucket.Get().(*lua.LState)
}

func (p *Pool) Coroutine() *lua.LState {
	co := p.bucket.Get().(*lua.LState)
	return co
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

func (p *Pool) Free(co *lua.LState) {
	co.Keepalive()
	p.bucket.Put(co)
}
