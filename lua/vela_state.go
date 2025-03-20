package lua

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

func (ls *LState) CheckObject(n int) LValue {
	lv := ls.Get(n)

	if lv.Type() != LTObject {
		ls.TypeError(n, LTObject)
		return nil
	}
	return lv
}

func (ls *LState) PushTo(v interface{}) {
	ls.Push(ReflectTo(v))
}

func (ls *LState) Pushf(format string, v ...interface{}) {
	ls.Push(LString(fmt.Sprintf(format, v...)))
}

func (ls *LState) CheckSocket(n int) string {
	v := ls.CheckString(n)
	if e := CheckSocket(v); e != nil {
		ls.RaiseError("must be socket , got fail , error:%v", e)
		return ""
	}
	return v

}

func (ls *LState) CheckSockets(n int) string {
	v := ls.CheckString(n)
	arr := strings.Split(v, ",")

	var err error
	for _, item := range arr {
		err = CheckSocket(item)
		if err != nil {
			ls.RaiseError("%s error: %v", err)
			return ""
		}
	}

	return v
}

func (ls *LState) CheckFile(n int) string {
	v := ls.CheckString(n)

	_, err := os.Stat(v)
	if os.IsNotExist(err) {
		ls.RaiseError("not found %s file", v)
		return ""
	}

	return v
}

func (ls *LState) IsTrue(n int) bool {
	return IsTrue(ls.Get(n))
}

func (ls *LState) IsFalse(n int) bool {
	return IsFalse(ls.Get(n))
}

func (ls *LState) IsNumber(n int) LNumber {
	return IsNumber(ls.Get(n))
}

func (ls *LState) IsInt(n int) int {
	return IsInt(ls.Get(n))
}

func (ls *LState) IsFunc(n int) *LFunction {
	return IsFunc(ls.Get(n))
}

func (ls *LState) IsString(n int) string {
	return IsString(ls.Get(n))
}

type CallBackFunction func(LValue) (stop bool)

func (ls *LState) Callback(fn CallBackFunction) {
	n := ls.GetTop()
	if n == 0 {
		return
	}

	for i := 1; i <= n; i++ {
		if fn(ls.Get(i)) {
			return
		}
	}
}

func (ls *LState) StackTrace(level int) string {
	return ls.stackTrace(level)
}

func (ls *LState) Exdata() any {
	return ls.private.Exdata
}

func (ls *LState) Exdata2() any {
	return ls.private.Exdata2
}

func (ls *LState) Cancel() {
	if ls.ctxCancelFn != nil {
		ls.ctxCancelFn()
	}
	return
}

func (ls *LState) Terminated() {
	if ls.private.Terminated == nil {
		return
	}
	ls.private.Terminated <- struct{}{}
	ls.RemoveCallerFrame()
}

func (ls *LState) NewThreadEx() *LState {
	options := ls.Options
	al := newAllocator(32)
	co := &LState{
		name:         ls.name,
		G:            ls.G,
		Env:          ls.Env,
		Parent:       nil,
		Panic:        panicWithTraceback,
		Dead:         false,
		Options:      options,
		stop:         0,
		alloc:        al,
		currentFrame: nil,
		wrapped:      false,
		uvcache:      nil,
		hasErrorFunc: false,
		mainLoop:     ls.mainLoop,
		ctx:          nil,
	}

	if options.MinimizeStackMemory {
		co.stack = newAutoGrowingCallFrameStack(options.CallStackSize)
	} else {
		co.stack = newFixedCallFrameStack(options.CallStackSize)
	}

	co.reg = newRegistry(co, options.RegistrySize, options.RegistryGrowStep, options.RegistryMaxSize, al)
	return co
}

func (ls *LState) pool() *sync.Pool {
	if ls.private.Pool == nil {
		ls.private.Pool = &sync.Pool{
			New: func() interface{} {
				return ls.NewThreadEx()
			},
		}
	}
	return ls.private.Pool
}

func (ls *LState) Coroutine() *LState {
	co := ls.pool().Get().(*LState)

	if ls.ctx != nil {
		co.mainLoop = mainLoopWithContext
		ctx, cancel := context.WithCancel(ls.ctx)
		co.ctx = ctx
		co.ctxCancelFn = cancel
	}

	co.private.Exdata = ls.private.Exdata
	co.private.Exdata2 = ls.private.Exdata2
	co.private.Pool = ls.private.Pool
	co.private.Terminated = make(chan struct{}, 1)
	return co
}

func (ls *LState) Keepalive(co *LState) {
	co.private.Exdata2 = nil
	co.private.Terminated = nil
	ls.pool().Put(co)
}

func (ls *LState) PanicErr(e error) {
	ls.RaiseError("%v", e)
}

func (ls *LState) Name() string {
	return ls.name
}

func NewStateEx(name string, fns ...func(*Options)) *LState {
	opt := &Options{
		CallStackSize: 128,
		RegistrySize:  128,
	}
	for _, fn := range fns {
		fn(opt)
	}

	if opt.CallStackSize < 1 {
		opt.CallStackSize = 256
	}

	if opt.CallStackSize < 128 {
		opt.RegistrySize = 128
	}

	if opt.RegistryMaxSize < opt.RegistrySize {
		opt.RegistryMaxSize = 0 // disable growth if max size is smaller than initial size
	} else {
		// if growth enabled, grow step is set
		if opt.RegistryGrowStep < 1 {
			opt.RegistryGrowStep = RegistryGrowStep
		}
	}

	co := newLState(*opt)
	if !opt.SkipOpenLibs {
		co.OpenLibs()
	}

	co.name = name
	co.private.Exdata = opt.Exdata
	co.private.Terminated = make(chan struct{}, 1)
	return co
}
