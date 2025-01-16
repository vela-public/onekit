package layer

import (
	"go.etcd.io/bbolt"
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

func DB() *bbolt.DB {
	if setting.Env == nil {
		panic("Environment is not Configured")
	}
	return setting.Env.DB()
}

func SHM() *bbolt.DB {
	if setting.Env == nil {
		panic("Environment is not Configured")
	}
	return setting.Env.SHM()
}

func Logger() LoggerType {
	if setting.Env == nil {
		panic("Environment is not Configured")
	}
	return setting.Env.Logger()
}

func ID() string {
	if setting.Env == nil {
		panic("Environment is not Configured")
	}
	return setting.Env.ID()
}

func IP() string {
	if setting.Env == nil {
		panic("Environment is not Configured")
	}
	return setting.Env.IP()
}

func Apply(env Environment) {
	setting.once.Do(func() {
		setting.Env = env
	})
}
