package luakit

import (
	"context"
	"github.com/vela-public/onekit/lua"
)

type Kit struct {
	U lua.UserKV            //local
	G map[string]lua.LValue //global
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

func (k *Kit) NewState(ctx context.Context, opts ...lua.Options) *lua.LState {
	co := lua.NewState(opts...)
	co.SetContext(ctx)
	co.PreloadModule("luakit", func(L *lua.LState) int {
		L.Push(k.U)
		return 1
	})

	co.SetGlobal("luakit", k.U)
	for name, value := range k.G {
		co.SetGlobal(name, value)
	}
	return co
}

func Apply(preloads ...func(lua.Preloader)) *Kit {
	kit := &Kit{
		U: lua.NewUserKV(),
	}
	builtin(kit.U)

	for _, loader := range preloads {
		loader(kit)
	}
	return kit
}
