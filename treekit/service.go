package treekit

import (
	"context"
	"fmt"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipe"
	"go.uber.org/zap/zapcore"
	"reflect"
	"strings"
	"sync"
	"time"
)

type MicroServiceType interface {
	Key() string
	Hash() string
	Metadata() libkit.DataKV[string, any]
}

type MicroService struct {
	//config
	config *MicoServiceConfig

	//root
	root *MsTree

	//setting
	setting struct {
		Level     zapcore.Level
		Keepalive bool
	}

	private struct {
		//with lua.LState context
		Context context.Context

		//cancel
		Cancel context.CancelFunc

		//uptime
		Uptime time.Time //uptime

		//error
		Error error
		Stack string

		//flag for MicroService
		mute sync.Mutex
		Flag SerErrNo

		//LState
		LState *lua.LState
	}

	//task processes element
	processes struct {
		mutex sync.RWMutex
		data  []*Process
		Link  []string //link other task
	}

	//event handle
	handler struct {
		After  *pipe.Chain
		Before *pipe.Chain
		Error  *pipe.Chain
		Audit  *pipe.Chain
	}
}

func (ms *MicroService) has(op SerErrNo) bool {
	ms.private.mute.Lock()
	defer ms.private.mute.Unlock()
	return (ms.private.Flag & op) == op
}

func (ms *MicroService) put(op SerErrNo) {
	ms.private.mute.Lock()
	defer ms.private.mute.Unlock()
	ms.private.Flag = ms.private.Flag | op
}

func (ms *MicroService) undo(op SerErrNo) {
	ms.private.mute.Lock()
	defer ms.private.mute.Unlock()

	ms.private.Flag &= ^op
}

func (ms *MicroService) set(op SerErrNo) {
	ms.private.mute.Lock()
	defer ms.private.mute.Unlock()

	ms.private.Flag = op
}

func (ms *MicroService) build() {
	parent := ms.root.Context()
	ctx, cancel := context.WithCancel(parent)

	//
	ms.private.Context = ctx
	ms.private.Cancel = cancel
	ms.handler.Before = pipe.NewChain()
	ms.handler.After = pipe.NewChain()
	ms.handler.Error = pipe.NewChain()
	ms.handler.Audit = pipe.NewChain()

	//init lua.LState coroutine
	kit := ms.root.LuaKit() // 功能的注入 lua 虚拟机
	ms.Preload(kit)
	ms.private.LState = kit.NewState(ctx, ms.Key(), func(option *lua.Options) {
		option.Exdata = ms
	})

	ms.set(Register)
}

func (ms *MicroService) update(config *MicoServiceConfig) {
	ms.config = config
	ms.set(Update)
}

func (ms *MicroService) reset() {
	sz := len(ms.processes.data)
	if sz == 0 {
		return
	}

	for i := 0; i < sz; i++ {
		ms.processes.data[i].put(Reset)
	}
}

func (ms *MicroService) clean() {
	sz := len(ms.processes.data)
	if sz == 0 {
		return
	}

	name := func(i int) string {
		return fmt.Sprintf("%s.%d", ms.config.Key, i)
	}

	errs := errkit.New()
	var curr []*Process
	push := func(srv *Process) {
		curr = append(curr, srv)
	}

	for i := 0; i < sz; i++ {
		p := ms.processes.data[i]
		key := name(i)

		if !p.has(Reset) {
			push(p)
			continue
		}

		if _ = p.Close(); p.info != nil {
			errs.Try(key, fmt.Errorf("%s  fail %v", p.Name(), p.info))
		} else {
			errs.Try(key, fmt.Errorf("%s succeed", p.Name()))
		}
	}

	if e := errs.Wrap(); e != nil {
		ms.NoError("%s", errs.Error())
	}
	ms.processes.data = curr
}

func (ms *MicroService) succeed() {
	ms.set(Running)
	ms.config.Source = nil
	ms.private.Error = nil
}

func (ms *MicroService) disable() {
	ms.set(Disable)
	ms.config.Source = nil
	ms.private.Error = nil
}

func (ms *MicroService) fail(err error) {
	ms.set(Fail)
	ms.private.Error = err
}

func (ms *MicroService) panic(err error) {
	ms.set(Panic)
	ms.private.Error = err
}

func (ms *MicroService) wakeup() error {
	if ms.has(Waking) {
		return fmt.Errorf("task waking up")
	} else {
		ms.put(Waking)
	}

	defer ms.undo(Waking)

	switch {
	case ms.has(Undefined):
		return fmt.Errorf("task undefined")
	case ms.has(Running):
		return nil
	case ms.has(Panic):
		//ms.OnError("task panic")
		return nil
	case ms.has(Disable):
		return nil
	case ms.has(Update):
		ms.reset()
		ms.SafeCall()
		ms.clean()
		return ms.UnwrapErr()
	case ms.has(Register):
		ms.SafeCall()
		return ms.UnwrapErr()
	case ms.has(Fail):
		ms.SafeCall()
		return ms.UnwrapErr()
	default:
		return fmt.Errorf("unknown option flag")
	}
}

func (ms *MicroService) Key() string {
	return ms.config.Key
}

func (ms *MicroService) ID() int64 {
	return ms.config.ID
}

func (ms *MicroService) Shutdown(s ProcessType, x func(error)) {
	srvName := ms.Key()
	name := s.Name()
	pro, exist := ms.have(name)
	if !exist {
		err := fmt.Errorf("%s not found %s", srvName, name)
		x(err)
		return
	}

	from := pro.From()
	if srvName != from {
		err := fmt.Errorf("%s processes.from=%s with %s not allow", srvName, pro.from, pro.Name())
		x(err)
		return
	}

	if ms.has(Disable) {
		err := fmt.Errorf("%s processes.from=%s disable", srvName, from)
		x(err)
		return
	}

	if pro.has(Succeed) {
		if err := pro.Close(); err != nil {
			pro.info = fmt.Errorf("%s close fail error %v", pro.Name(), err)
			pro.set(Failed)
			x(pro.info)
			return
		} else {
			pro.info = nil
			pro.set(Stopped)
		}
		return
	}

	pro.info = fmt.Errorf("%s close fail error not running", pro.Name())
	return
}

func (ms *MicroService) Startup(s ProcessType, x func(error)) {
	srvName := ms.Key()
	name := s.Name()
	pro, exist := ms.have(name)
	if !exist {
		err := fmt.Errorf("%s not found %s", srvName, name)
		x(err)
		return
	}

	from := pro.From()
	if srvName != from {
		err := fmt.Errorf("%s processes.from=%s with %s not allow", srvName, pro.from, pro.Name())
		x(err)
		return
	}

	if ms.has(Disable) {
		err := fmt.Errorf("%s processes.from=%s disable", srvName, from)
		x(err)
		return
	}

	switch {
	case pro.has(Succeed):
		if rd, ok := pro.data.(ReloadType); ok {
			err := rd.Reload()
			if err != nil {
				pro.info = err
				pro.set(Failed)
				x(err)
			} else {
				pro.info = nil
				pro.set(Succeed)
			}
			return
		}

		if err := pro.Close(); err != nil {
			pro.info = fmt.Errorf("%s close fail error %v", pro.Name(), err)
			pro.set(Failed)
			x(pro.info)
			return
		} else {
			pro.info = nil
			pro.set(Stopped)
		}

		if err := pro.data.Start(); err != nil {
			pro.info = fmt.Errorf("%s open fail error %v", pro.Name(), err)
			pro.set(Failed)
			x(pro.info)
		} else {
			pro.set(Succeed)
			pro.info = nil
		}

	default:
		if err := pro.data.Start(); err != nil {
			pro.info = fmt.Errorf("%s open fail error %v", pro.Name(), err)
			pro.set(Failed)
			x(pro.info)
		} else {
			pro.info = nil
			pro.set(Succeed)
		}
	}
}

func (ms *MicroService) verify(L *lua.LState) error {
	dat := L.Exdata()
	t2, ok := dat.(*MicroService)
	if !ok {
		return fmt.Errorf("not task vm")
	}

	if t2.Key() != ms.Key() {
		return fmt.Errorf("mismatch task must %s got %s", ms.Key(), t2.Key())
	}

	return nil
}

func (ms *MicroService) define(name string, typeof string) *Process {
	ms.processes.mutex.Lock()
	defer ms.processes.mutex.Unlock()

	pro := &Process{
		name:   name,
		flag:   defined,
		from:   ms.Key(),
		typeof: typeof,
	}

	ms.processes.data = append(ms.processes.data, pro)
	return pro
}

func (ms *MicroService) have(name string) (*Process, bool) {
	ms.processes.mutex.RLock()
	defer ms.processes.mutex.RUnlock()

	sz := len(ms.processes.data)
	var pro *Process
	for i := 0; i < sz; i++ {
		pro = ms.processes.data[i]
		if pro.name == name {
			return pro, true
		}
	}
	return nil, false
}

func (ms *MicroService) Create(L *lua.LState, name string, typeof string) *Process {

	sz := len(ms.processes.data)
	var pro *Process
	for i := 0; i < sz; i++ {
		p := ms.processes.data[i]
		if p.name == name {
			pro = p
			break
		}
	}

	if pro == nil {
		return ms.define(name, typeof)
	}

	if typ := reflect.TypeOf(pro.data).String(); typ != typeof {
		L.RaiseError("not allow change type , must %s but got %s", typeof, typ)
		return pro
	}

	return pro
}

func (ms *MicroService) SafeCall() *MicroService {
	fn, err := ms.private.LState.Load(ms.config.NewReader(), ms.config.Key)
	if err != nil {
		ms.panic(err)
		return ms
	}

	ms.private.Uptime = time.Now()
	err = ms.private.LState.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: ms.root.Protect(),
	})

	if err == nil {
		ms.succeed()
		ms.root.Debugf("service.%s succeed", ms.Key())
		return ms
	}

	if e, ok := err.(*lua.ApiError); ok && strings.HasSuffix(e.Object.String(), "disable") {
		ms.disable()
		ms.root.Debugf("service.%s disable", ms.Key())
	} else {
		ms.fail(err)
		ms.root.Debugf("service.%s fail %v", ms.Key(), err)
	}
	return ms
}

func (ms *MicroService) Enable(v zapcore.Level) bool {
	return v >= ms.setting.Level
}

func (ms *MicroService) UnwrapErr() error {
	return ms.private.Error
}

func (ms *MicroService) Done() <-chan struct{} {
	return ms.private.Context.Done()
}

func (ms *MicroService) Close() {
	ms.processes.mutex.Lock()
	defer ms.processes.mutex.Unlock()

	sz := len(ms.processes.data)
	//ev := event.Lazy().Trace("service")
	//ev.FromCode = ms.config.Key
	//ev.Subject = "关闭服务"
	//ev.Message = fmt.Sprintf("关闭了%d工作进程", sz)
	for i := 0; i < sz; i++ {
		pro := ms.processes.data[i]
		_ = pro.Close()
		//ev.Metadata[pro.name] = pro.UnwrapErr()
	}
	//ev.Error().Report()
}
