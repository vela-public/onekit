package layer

import (
	"sync"
)

var setting = struct {
	once sync.Once
	Env  Environment
}{}

func LazyEnv() Environment {
	if setting.Env == nil {
		panic("Environment is not Configured")
	}

	return setting.Env
}

func Apply(env Environment) {
	setting.once.Do(func() {
		setting.Env = env
	})
}
