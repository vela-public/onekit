package luakit

import (
	"context"
	"github.com/vela-public/onekit/lua"
)

type Kit struct {
	name string
	U    lua.UserKV            //local
	G    map[string]lua.LValue //global
}

func (k *Kit) Clone() *Kit {
	k2 := &Kit{
		name: k.name,
		U:    lua.NewUserKV(),
		G:    make(map[string]lua.LValue),
	}

	k.U.ForEach(func(key string, val lua.LValue) bool {
		k2.U.Set(key, val)
		return true
	})

	for key, val := range k.G {
		k2.G[key] = val
	}
	return k2
}

func (k *Kit) Name() string {
	return k.name
}

func (k *Kit) Set(s string, value lua.LValue) {
	k.U.Set(s, value)
}

func (k *Kit) SetGlobal(s string, value lua.LValue) {
	if k.G == nil {
		k.G = make(map[string]lua.LValue)
	}
	k.G[s] = value
}

func (k *Kit) Get(s string) lua.LValue {
	return k.U.Get(s)
}

func (k *Kit) Global(s string) lua.LValue {
	if k.G == nil {
		return lua.LNil
	}

	return k.G[s]
}

func (k *Kit) NewState(ctx context.Context, options ...func(*lua.Options)) *lua.LState {
	co := lua.NewStateEx(options...)
	co.SetContext(ctx)

	//local luakit = require("luakit")
	co.PreloadModule(k.name, func(L *lua.LState) int {
		L.Push(k.U)
		return 1
	})

	//luakit = luakit.elastic
	co.SetGlobal(k.name, k.U)
	for name, value := range k.G {
		co.SetGlobal(name, value)
	}
	return co
}

func Apply(name string, preloads ...func(lua.Preloader)) *Kit {
	kit := &Kit{
		name: name,
		U:    lua.NewUserKV(),
	}
	builtin(kit.U)

	for _, loader := range preloads {
		loader(kit)
	}
	return kit
}
