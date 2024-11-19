package filekit

import (
	"github.com/vela-public/onekit/pipe"
	"go.etcd.io/bbolt"
)

type FileTailFunc func(*FileTail)
type FileTailGuide struct{}

func NewTailGuide() FileTailGuide {
	return FileTailGuide{}
}

var Tail = FileTailGuide{}

func (FileTailGuide) FastJSON() FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.FastJSON = true
	}
}

func (FileTailGuide) Log(l Logger) FileTailFunc {
	return func(ft *FileTail) {
		ft.logger = l
	}
}

func (FileTailGuide) Limit(n int) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Limit = n
	}
}

func (FileTailGuide) Follow(b bool) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Follow = b
	}
}

func (FileTailGuide) Delim(byt byte) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Delim = byt
	}
}

func (FileTailGuide) Pipe(v any, options ...func(*pipe.HandleEnv)) FileTailFunc {
	return func(ft *FileTail) {
		ft.private.Chain.NewHandler(v, options...)
	}
}

func (FileTailGuide) Db(b *bbolt.DB) FileTailFunc {
	return func(ft *FileTail) {
		ft.private.DB = b
	}
}

func (FileTailGuide) SkipFile(fn func(string) bool) FileTailFunc {
	return func(ft *FileTail) {
		ft.private.SkipFile = append(ft.private.SkipFile, fn)
	}
}

func (FileTailGuide) WaitFor(n int) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Wait = n
	}
}

func (FileTailGuide) Location(offset int64, whence int) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Mode = "location"
		ft.setting.Location.Offset = offset
		ft.setting.Location.Whence = whence
	}

}

func (FileTailGuide) Thread(n int) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Thread = n
	}
}

func (FileTailGuide) Target(s ...string) FileTailFunc {
	return func(ft *FileTail) {
		ft.setting.Target = append(ft.setting.Target, s...)
	}
}
