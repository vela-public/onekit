package filekit

import (
	"github.com/vela-public/onekit/pipe"
	"go.etcd.io/bbolt"
)

type FileTailFunc func(*FileTail)
type LazyFileTail struct{}

func LazyTail() LazyFileTail {
	return LazyFileTail{}
}

func (LazyFileTail) FastJSON() FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.FastJSON = true
	}
}

func (LazyFileTail) Log(l Logger) FileTailFunc {
	return func(ft *FileTail) {
		ft.logger = l
	}
}

func (LazyFileTail) Limit(n int) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Limit = n
	}
}

func (LazyFileTail) Follow(b bool) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Follow = b
	}
}

func (LazyFileTail) Delim(byt byte) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Delim = byt
	}
}

func (LazyFileTail) Pipe(v any, options ...func(*pipe.HandleEnv)) FileTailFunc {
	return func(ft *FileTail) {
		ft.private.Chain.NewHandler(v, options...)
	}
}

func (LazyFileTail) Db(b *bbolt.DB) FileTailFunc {
	return func(ft *FileTail) {
		ft.private.DB = b
	}
}

func (LazyFileTail) SkipFile(fn func(string) bool) FileTailFunc {
	return func(ft *FileTail) {
		ft.private.SkipFile = append(ft.private.SkipFile, fn)
	}
}

func (LazyFileTail) WaitFor(n int) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Wait = n
	}
}

func (LazyFileTail) Location(offset int64, whence int) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Mode = "location"
		ft.setting.Location.Offset = offset
		ft.setting.Location.Whence = whence
	}

}

func (LazyFileTail) Thread(n int) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Thread = n
	}
}

func (LazyFileTail) Target(s ...string) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Target = append(ft.setting.Target, s...)
	}
}
