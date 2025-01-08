package layer

import (
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"sync"
)

var setting = struct {
	once sync.Once
	Kit  *luakit.Kit
	Env  Environment
}{}

func LazyEnv() Environment {
	if setting.Env == nil {
		panic("Environment is not Configured")
	}

	return setting.Env
}

func Luakit() *luakit.Kit {
	if setting.Kit == nil {
		panic("Lua is not Configured")
	}
	return setting.Kit
}

func Apply(env Environment, options ...func(lua.Preloader)) {
	setting.once.Do(func() {
		setting.Env = env
		setting.Kit = luakit.Apply(env.Name(), options...)
	})
}
