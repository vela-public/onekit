package filekit

import (
	"context"
	"github.com/vela-public/onekit/pipekit"
	"go.etcd.io/bbolt"
	"io"
	"strings"
)

type FileTailFunc func(*FileTail)
type LazyFileTail struct {
	ctx     context.Context
	pattern []string
	tail    *FileTail
	err     error
}

func LazyTail(ctx context.Context, pattern ...string) *LazyFileTail {
	name := strings.Join(pattern, ",")
	tail := &LazyFileTail{
		ctx:     ctx,
		pattern: pattern,
		tail:    NewTail(name),
	}
	tail.Target(pattern...)
	return tail
}

func (l *LazyFileTail) FastJSON() *LazyFileTail {
	l.tail.setting.FastJSON = true
	return l
}

func (l *LazyFileTail) Log(log Logger) *LazyFileTail {
	l.tail.logger = log
	return l
}

func (l *LazyFileTail) Limit(n int) *LazyFileTail {
	l.tail.setting.Limit = n
	return l
}

func (l *LazyFileTail) Follow(b bool) *LazyFileTail {
	l.tail.setting.Follow = b
	return l
}

func (l *LazyFileTail) Delim(byt byte) *LazyFileTail {
	l.tail.setting.Delim = byt
	return l
}

func (l *LazyFileTail) Pipe(v any, options ...func(*pipekit.HandleEnv)) *LazyFileTail {
	l.tail.private.Chain.NewHandler(v, options...)
	return l
}

func (l *LazyFileTail) Output(v any, options ...func(*pipekit.HandleEnv)) *LazyFileTail {
	l.tail.private.Chain.NewHandler(v, options...)
	l.err = l.tail.Background(l.ctx)
	return l
}

func (l *LazyFileTail) Db(db *bbolt.DB) *LazyFileTail {
	l.tail.private.Seeker = NewSeekDB(db, "SHM_FILE_RECORD")
	return l
}
func (l *LazyFileTail) SeekEnd() *LazyFileTail {
	l.Location(0, io.SeekEnd)
	return l
}

func (l *LazyFileTail) Mem() *LazyFileTail {
	l.tail.private.Seeker = NewSeekMem()
	return l
}

func (l *LazyFileTail) SkipFile(fn func(string) bool) *LazyFileTail {
	l.tail.private.SkipFile = append(l.tail.private.SkipFile, fn)
	return l
}

func (l *LazyFileTail) WaitFor(n int) *LazyFileTail {
	l.tail.setting.Wait = n
	return l
}

func (l *LazyFileTail) Location(offset int64, whence int) *LazyFileTail {
	l.tail.setting.Mode = "location"
	l.tail.setting.Location.Offset = offset
	l.tail.setting.Location.Whence = whence
	return l
}

func (l *LazyFileTail) Thread(n int) *LazyFileTail {
	l.tail.setting.Thread = n
	return l
}

func (l *LazyFileTail) Target(s ...string) {
	l.tail.setting.Target = append(l.tail.setting.Target, s...)
}

func (l *LazyFileTail) Unwrap() *FileTail {
	return l.tail
}

func (l *LazyFileTail) UnwrapErr() error {
	return l.err
}
