package pipe

import (
	"github.com/vela-public/onekit/lua"
	"sync"
)

var DefaultLuaThreadPool = NewLuaThreadPool(lua.NewState())

type LuaThreadPool struct {
	Main *lua.LState
	Pool *sync.Pool
}

func (l *LuaThreadPool) Coroutine() *lua.LState {
	co, _ := l.Main.NewThread()
	return co
}

func (l *LuaThreadPool) Get() *lua.LState {
	return l.Pool.Get().(*lua.LState)
}

func (l *LuaThreadPool) Clone(parent *lua.LState) *lua.LState {
	co := l.Get()
	co.Copy(parent)
	return co
}

func (l *LuaThreadPool) Put(co *lua.LState) {
	co.Keepalive()
	l.Pool.Put(co)
}

func NewLuaThreadPool(L *lua.LState) *LuaThreadPool {
	pool := &sync.Pool{
		New: func() interface{} {
			co, _ := L.NewThread()
			return co
		},
	}

	return &LuaThreadPool{
		Main: L,
		Pool: pool,
	}
}
