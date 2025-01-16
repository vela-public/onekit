package taskit

import (
	"context"
	"fmt"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipe"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	Undefined TaskNo = 1 << iota
	Register
	Waking
	Running
	Panic
	Fail
	Update
	Disable
	Empty
)

var TaskNoMap = map[TaskNo]string{
	Undefined: "undefined",
	Register:  "register",
	Waking:    "waking",
	Running:   "running",
	Panic:     "panic",
	Fail:      "fail",
	Update:    "update",
	Disable:   "disable",
	Empty:     "empty",
}

func (tn TaskNo) String() string {
	return TaskNoMap[tn]
}

type TaskNo uint32

type task struct {
	//config
	config *Config

	//root
	root *Tree

	//setting
	setting struct {
		Debug     bool
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

		//flag for task
		mute sync.Mutex
		Flag TaskNo

		//LState
		LState *lua.LState
	}

	//task service element
	service struct {
		mutex sync.RWMutex
		store []*Service
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

func (t *task) has(op TaskNo) bool {
	t.private.mute.Lock()
	defer t.private.mute.Unlock()
	return (t.private.Flag & op) == op
}

func (t *task) put(op TaskNo) {
	t.private.mute.Lock()
	defer t.private.mute.Unlock()
	t.private.Flag = t.private.Flag | op
}

func (t *task) undo(op TaskNo) {
	t.private.mute.Lock()
	defer t.private.mute.Unlock()

	t.private.Flag &= ^op
}

func (t *task) set(op TaskNo) {
	t.private.mute.Lock()
	defer t.private.mute.Unlock()

	t.private.Flag = op
}

func (t *task) build() {
	parent := t.root.Context()
	ctx, cancel := context.WithCancel(parent)

	//
	t.private.Context = ctx
	t.private.Cancel = cancel
	t.handler.Before = pipe.NewChain()
	t.handler.After = pipe.NewChain()
	t.handler.Error = pipe.NewChain()
	t.handler.Audit = pipe.NewChain()

	//init lua.LState coroutine
	kit := t.root.NewKit() // 功能的注入 lua 虚拟机
	t.Preload(kit)
	t.private.LState = kit.NewState(ctx, t.Key(), func(option *lua.Options) {
		option.Payload = t
	})

	t.set(Register)
}

func (t *task) update(config *Config) {
	t.config = config
	t.set(Update)
}

func (t *task) reset() {
	sz := len(t.service.store)
	if sz == 0 {
		return
	}

	for i := 0; i < sz; i++ {
		t.service.store[i].put(Reset)
	}
}

func (t *task) clean() {
	sz := len(t.service.store)
	if sz == 0 {
		return
	}

	name := func(i int) string {
		return fmt.Sprintf("%s.%d", t.config.Key, i)
	}

	errs := errkit.New()
	var curr []*Service
	push := func(srv *Service) {
		curr = append(curr, srv)
	}

	for i := 0; i < sz; i++ {
		srv := t.service.store[i]
		key := name(i)

		if !srv.has(Reset) {
			push(srv)
			continue
		}

		if _ = srv.Close(); srv.info != nil {
			errs.Try(key, fmt.Errorf("%s  fail %v", srv.name, srv.info))
		} else {
			errs.Try(key, fmt.Errorf("%s succeed", srv.name))
		}
	}

	if e := errs.Wrap(); e != nil {
		t.NoError("%s", errs.Error())
	}
	t.service.store = curr
}

func (t *task) succeed() {
	t.set(Running)
	t.config.Source = nil
	t.private.Error = nil
}

func (t *task) disable() {
	t.set(Disable)
	t.config.Source = nil
	t.private.Error = nil
}

func (t *task) fail(err error) {
	t.set(Fail)
	t.private.Error = err
}

func (t *task) panic(err error) {
	t.set(Panic)
	t.private.Error = err
}

func (t *task) wakeup() error {
	if t.has(Waking) {
		return fmt.Errorf("task waking up")
	} else {
		t.put(Waking)
	}

	defer t.undo(Waking)

	switch {
	case t.has(Undefined):
		return fmt.Errorf("task undefined")
	case t.has(Running):
		return nil
	case t.has(Panic):
		//t.OnError("task panic")
		return nil
	case t.has(Disable):
		return nil
	case t.has(Update):
		t.reset()
		t.SafeCall()
		t.clean()
		return t.UnwrapErr()
	case t.has(Register):
		t.SafeCall()
		return t.UnwrapErr()
	case t.has(Fail):
		t.SafeCall()
		return t.UnwrapErr()
	default:
		return fmt.Errorf("unknown option flag")
	}
}

func (t *task) Key() string {
	return t.config.Key
}

func (t *task) ID() int64 {
	return t.config.ID
}

func (t *task) Do(s ServiceType, x func(error)) {
	key := t.Key()
	name := s.Name()
	srv, exist := t.have(name)
	if !exist {
		err := fmt.Errorf("current.task=%s not found %s", key, name)
		x(err)
		return
	}

	from := srv.From()
	if key != from {
		err := fmt.Errorf("current.task=%s service.from=%s with %s not allow", key, srv.from, srv.name)
		x(err)
		return
	}

	if t.has(Disable) {
		err := fmt.Errorf("current.task=%s service.from=%s disable", key, from)
		x(err)
		return
	}

	switch {
	case srv.has(Succeed):
		if rd, ok := srv.data.(ReloadType); ok {
			err := rd.Reload()
			if err != nil {
				srv.info = err
				srv.set(Failed)
				x(err)
			} else {
				srv.info = nil
				srv.set(Succeed)
			}
			return
		}

		if err := srv.Close(); err != nil {
			srv.info = fmt.Errorf("%s close fail error %v", srv.name, err)
			srv.set(Failed)
			x(srv.info)
			return
		} else {
			srv.info = nil
			srv.set(Stopped)
		}

		if err := srv.data.Start(); err != nil {
			srv.info = fmt.Errorf("%s open fail error %v", srv.name, err)
			srv.set(Failed)
			x(srv.info)
		} else {
			srv.set(Succeed)
			srv.info = nil
		}

	default:
		if err := srv.data.Start(); err != nil {
			srv.info = fmt.Errorf("%s open fail error %v", srv.name, err)
			srv.set(Failed)
			x(srv.info)
		} else {
			srv.info = nil
			srv.set(Succeed)
		}
	}
}

func (t *task) verify(L *lua.LState) error {
	dat := L.Payload()
	t2, ok := dat.(*task)
	if !ok {
		return fmt.Errorf("not task vm")
	}

	if t2.Key() != t.Key() {
		return fmt.Errorf("mismatch task must %s got %s", t.Key(), t2.Key())
	}

	return nil
}

func (t *task) define(name string) *Service {
	t.service.mutex.Lock()
	defer t.service.mutex.Unlock()

	srv := &Service{
		name: name,
		flag: defined,
		from: t.Key(),
	}

	t.service.store = append(t.service.store, srv)
	return srv
}

func (t *task) have(name string) (*Service, bool) {
	t.service.mutex.RLock()
	defer t.service.mutex.RUnlock()

	sz := len(t.service.store)
	var srv *Service
	for i := 0; i < sz; i++ {
		if srv = t.service.store[i]; srv.name == name {
			return srv, true

		}
	}
	return nil, false
}

func (t *task) Create(L *lua.LState, name string, typeof string) *Service {

	sz := len(t.service.store)
	var srv *Service
	for i := 0; i < sz; i++ {
		s := t.service.store[i]
		if s.name == name {
			srv = s
			break
		}
	}

	if srv == nil {
		return t.define(name)
	}

	if typ := reflect.TypeOf(srv.data).String(); typ != typeof {
		L.RaiseError("not allow change type , must %s but got %s", typeof, typ)
		return srv
	}

	return srv
}

func (t *task) SafeCall() *task {
	fn, err := t.private.LState.Load(t.config.NewReader(), t.config.Key)
	if err != nil {
		t.panic(err)
		return t
	}

	t.private.Uptime = time.Now()
	err = t.private.LState.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: t.root.Protect(),
	})

	if err == nil {
		t.succeed()
		return t
	}

	if e, ok := err.(*lua.ApiError); ok && strings.HasSuffix(e.Object.String(), "disable") {
		t.disable()
	} else {
		t.fail(err)
	}
	return t
}

func (t *task) UnwrapErr() error {
	return t.private.Error
}

func (t *task) Done() <-chan struct{} {
	return t.private.Context.Done()
}

func (t *task) Close() error {
	return nil
}

func (t *task) Start() error {
	return nil
}
