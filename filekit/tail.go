package filekit

import (
	"context"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/cond"
	"github.com/vela-public/onekit/gopool"
	"github.com/vela-public/onekit/jsonkit"
	"github.com/vela-public/onekit/noop"
	"github.com/vela-public/onekit/pipe"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

/*
	tail := &parent{
		buffer: 4096,
		delim: "\n",
	}
	tail.rate(100)
	tail.pipe( kfk , syslog , http )
	tail.start()
*/

type FileTail struct {
	logger  Logger
	setting *Setting

	datalog struct {
		read int64
		done int64
	}

	private struct {
		history  map[string]*Section
		limit    *limit
		context  context.Context
		cancel   context.CancelFunc
		queue    *gopool.Queue[*Line]
		Chain    *pipe.Chain
		Switch   *pipe.Switch
		Debug    *pipe.Chain
		Drop     *cond.Ignore
		SkipFile []func(string) bool
		Seeker   Seeker
	}
}

func (ft *FileTail) Name() string {
	return ft.setting.Name
}

func (ft *FileTail) Decode(line *Line) {
	f := &jsonkit.FastJSON{}
	f.ParseText(cast.B2S(line.Text))
	line.Json = f
}

func (ft *FileTail) Debugf(format string, v ...interface{}) {
	if ft.private.Debug.Len() == 0 {
		return
	}
	ft.private.Debug.Invoke(fmt.Sprintf(format, v...))
}
func (ft *FileTail) Errorf(format string, v ...interface{}) {
	ft.logger.Errorf(format, v...)
}

func (ft *FileTail) DataLog() (read, done int64) {
	return atomic.LoadInt64(&ft.datalog.read), atomic.LoadInt64(&ft.datalog.done)
}

func (ft *FileTail) Input(line *Line) {
	atomic.AddInt64(&ft.datalog.read, 1)

	if ft.setting.FastJSON {
		ft.Decode(line)
	}

	if ft.private.Drop.Match(line) {
		return
	}

	ft.private.queue.Push(line)
}

func (ft *FileTail) Wait() {
	if ft.private.limit != nil {
		ft.private.limit.wait()
	}
}

func (ft *FileTail) Tell(file string, offset int64) {
	if ft.private.Seeker == nil {
		return
	}

	err := ft.private.Seeker.Save(file, offset)
	if err != nil {
		ft.Errorf("%s tail save seek fail %v", file, err)
		return
	}

	ft.Debugf("%s tail save seek:%d", file, offset)
}

func (ft *FileTail) SeekTo(name string) int64 {
	if ft.private.Seeker == nil {
		return 0
	}
	seek, err := ft.private.Seeker.Find(name)
	if err != nil {
		ft.Errorf("%s tail seek fail %v", name, err)
	}
	return seek
}

func (ft *FileTail) WaitFor(callback func() (stop bool)) {

	//是否开启等待
	if ft.setting.Wait <= 0 { //防止频繁 操作句柄
		callback()
		return
	}

	//首次打开 无需等待
	if stop := callback(); stop {
		return
	}

	tk := time.NewTicker(time.Duration(ft.setting.Wait) * time.Second)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			if stop := callback(); stop {
				return
			}
		case <-ft.Done():
			ft.Errorf("%s wait exit", ft.Name())
			return
		}
	}
}

func (ft *FileTail) invoke(line *Line) {
	ft.private.Chain.Invoke(line)
	ft.private.Switch.Invoke(line)
}

func (ft *FileTail) Prepare(parent context.Context) {

	//初始化context
	ft.private.context,
		ft.private.cancel = context.WithCancel(parent)

	ft.private.limit = NewLimit(ft.private.context, ft.setting.Limit)
	ft.private.history = make(map[string]*Section)

	queue := gopool.NewQueue[*Line](ft.private.context, gopool.Workers(ft.setting.Thread), gopool.Ticker(5))
	queue.Handler(ft.invoke)
	queue.SetErrHandler(func(err error) {
		ft.Errorf(err.Error())
	})

	ft.private.queue = queue

}

func (ft *FileTail) clean(data map[string]*Section) {
	for filename, s := range data {
		if s.flag == Running { // Prevent slow reading speed
			ft.private.history[filename] = s
			continue
		}

		s.close()
		s.flag = Cleaned
		ft.Errorf("clean %s", s.path)
	}
}

func (ft *FileTail) SkipFile(filename string) bool {
	if len(ft.private.SkipFile) == 0 {
		return false
	}

	for _, skip := range ft.private.SkipFile {
		if skip(filename) {
			return true
		}
	}
	return false
}
func (ft *FileTail) GlobFor() {
	history := make(map[string]*Section)
	sz := len(ft.setting.Target)
	if sz == 0 {
		return
	}

	for i := 0; i < sz; i++ {
		pattern := ft.setting.Target[i]
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range files {
			if ft.SkipFile(file) {
				continue
			}
			ft.Detect(history, file)
		}
	}

	ft.clean(ft.private.history)
	ft.private.history = history
}

func (ft *FileTail) Detect(history map[string]*Section, file string) {
	filename, err := filepath.Abs(file)
	if err != nil {
		ft.Errorf("detect %s fail %v", filename, err)
		return
	}

	stat, err := os.Stat(filename)
	if err != nil {
		ft.Errorf("read stat %s fail %v", filename, err)
		return
	}

	if stat.IsDir() {
		return
	}

	s, ok := ft.private.history[filename]
	if !ok {
		s = &Section{
			tail: ft,
			path: filename,
		}
		history[filename] = s
	} else {
		history[filename] = s
		delete(ft.private.history, filename)
	}
	s.detect()
}

func (ft *FileTail) scanner() {
	tk := time.NewTicker(time.Duration(ft.setting.Poll) * time.Second)
	defer tk.Stop()

	//首次打开 无需等待
	ft.GlobFor()

	for {
		select {
		case <-ft.Done():
			ft.Errorf("%s watch exit", ft.Name())
			return
		case <-tk.C:
			ft.GlobFor()
		}
	}
}

func (ft *FileTail) Background(ctx context.Context) error {
	if err := ft.setting.Bad(); err != nil {
		return err
	}

	ft.Prepare(ctx)
	go ft.scanner()
	return nil
}

func (ft *FileTail) Run(ctx context.Context) error {
	if err := ft.setting.Bad(); err != nil {
		return err
	}

	ft.Prepare(ctx)
	ft.scanner()
	return nil
}

func (ft *FileTail) Done() <-chan struct{} {
	return ft.private.context.Done()
}

func (ft *FileTail) Cancel() {
	ft.private.cancel()
}

func (ft *FileTail) Close() error {
	ft.Cancel()
	return nil
}

func (ft *FileTail) Switch() *pipe.Switch {
	return ft.private.Switch
}

func (ft *FileTail) Chain() *pipe.Chain {
	return ft.private.Chain
}

func (ft *FileTail) Drop() *cond.Ignore {
	return ft.private.Drop
}

func (ft *FileTail) Apply(opts ...FileTailFunc) {
	for _, fn := range opts {
		fn(ft)
	}
}

func NewTail(name string, opts ...FileTailFunc) *FileTail {
	cfg := Default(name)
	ft := &FileTail{
		setting: cfg,
		logger:  noop.NewLogger(2),
	}
	for _, fn := range opts {
		fn(ft)
	}

	ft.private.Chain = pipe.NewChain()
	ft.private.Switch = pipe.NewSwitch()
	ft.private.Debug = pipe.NewChain()
	ft.private.Drop = cond.NewIgnore()
	return ft
}
