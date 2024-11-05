package filekit

import (
	"github.com/vela-public/onekit/pipe"
	"go.etcd.io/bbolt"
)

func FastJSON() func(*FileTail) {
	return func(ft *FileTail) {
		ft.setting.FastJSON = true
	}
}

func Log(l Logger) func(*FileTail) {
	return func(ft *FileTail) {
		ft.logger = l
	}
}

func Limit(n int) func(*FileTail) {
	return func(ft *FileTail) {
		ft.setting.Limit = n
	}
}

func Follow(b bool) func(*FileTail) {
	return func(ft *FileTail) {
		ft.setting.Follow = b
	}
}

func Delim(byt byte) func(*FileTail) {
	return func(ft *FileTail) {
		ft.setting.Delim = byt
	}
}

func Pipe(v any, options ...func(*pipe.HandleEnv)) func(*FileTail) {
	return func(ft *FileTail) {
		ft.private.Chain.NewHandler(v, options...)
	}
}

func Db(b *bbolt.DB) func(*FileTail) {
	return func(ft *FileTail) {
		ft.private.DB = b
	}
}

func WaitFor(n int) func(*FileTail) {
	return func(ft *FileTail) {
		ft.setting.Wait = n
	}
}

func Thread(n int) func(*FileTail) {
	return func(ft *FileTail) {
		ft.setting.Thread = n
	}
}

func Target(s ...string) func(*FileTail) {
	return func(ft *FileTail) {
		ft.setting.Target = append(ft.setting.Target, s...)
	}
}
