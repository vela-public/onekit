package treekit

import (
	"context"
	"fmt"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"go.uber.org/zap/zapcore"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Reply struct {
	ID      int64                       `json:"id"`
	ExecID  int64                       `json:"exec_id"`
	Succeed bool                        `json:"succeed"`
	Data    *libkit.DataKV[string, any] `json:"data"`
	Reason  string                      `json:"reason"`
}

type Task struct {
	//config
	config *TaskConfig
	reply  *Reply

	//setting
	setting struct {
		Level zapcore.Level
	}

	private struct {
		Context context.Context
		Cancel  context.CancelFunc
		Tree    *TaskTree
		LState  *lua.LState
	}

	processes struct {
		mutex sync.RWMutex
		data  []*Process
		Link  []string //link other task
	}
}

func (t *Task) Tree() *TaskTree {
	return t.private.Tree
}

func (t *Task) Reply() *Reply {
	return t.reply
}

func (t *TaskTree) NewKit() *luakit.Kit {
	return t.private.luakit.Clone()
}

func (t *Task) From() string {
	return t.private.LState.Name()
}

func (t *Task) Timeout() time.Duration {
	if t.config.Timeout == 0 {
		return 300 * time.Second
	}
	return time.Duration(t.config.Timeout) * time.Second
}

func (t *Task) Key() string {
	return fmt.Sprintf("task.%s", t.config.Name)
}

func (t *Task) define(name string) *Process {
	t.processes.mutex.Lock()
	defer t.processes.mutex.Unlock()

	pro := &Process{
		name: name,
		flag: defined,
		from: t.Key(),
	}

	t.processes.data = append(t.processes.data, pro)
	return pro
}

func (t *Task) Create(L *lua.LState, name string, typeof string) *Process {
	sz := len(t.processes.data)
	var pro *Process
	for i := 0; i < sz; i++ {
		p := t.processes.data[i]
		if p.name == name {
			pro = p
			break
		}
	}

	if pro == nil {
		return t.define(name)
	}

	if typ := reflect.TypeOf(pro.data).String(); typ != typeof {
		L.RaiseError("not allow change type , must %s but got %s", typeof, typ)
		return pro
	}

	return pro
}

func (t *Task) have(name string) (*Process, bool) {
	t.processes.mutex.RLock()
	defer t.processes.mutex.RUnlock()
	sz := len(t.processes.data)
	for i := 0; i < sz; i++ {
		if p := t.processes.data[i]; p.name == name {
			return p, true
		}
	}
	return nil, false
}

func (t *Task) Shutdown(v ProcessType, x func(error)) {
	key := t.Key()
	name := v.Name()
	pro, exist := t.have(name)
	if !exist {
		err := fmt.Errorf("%s not found %s", key, name)
		x(err)
		return
	}

	from := pro.From()
	if key != from {
		err := fmt.Errorf("%s processes.from=%s with %s not allow", key, pro.from, pro.name)
		x(err)
		return
	}

	if pro.has(Succeed) {
		if err := pro.Close(); err != nil {
			pro.info = fmt.Errorf("%s close fail error %v", pro.name, err)
			pro.set(Failed)
			x(pro.info)
			return
		} else {
			pro.info = nil
			pro.set(Stopped)
		}
		return
	}

	pro.info = fmt.Errorf("%s close fail error not running", pro.name)
	return
}

func (t *Task) Startup(ctx context.Context, v ProcessType, x func(error)) {
	key := t.Key()
	name := v.Name()
	pro, exist := t.have(name)
	if !exist {
		err := fmt.Errorf("%s not found %s", key, name)
		x(err)
		return
	}

	from := pro.From()
	if key != from {
		err := fmt.Errorf("%s processes.from=%s with %s not allow", key, pro.from, pro.name)
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
			pro.info = fmt.Errorf("%s close fail error %v", pro.name, err)
			pro.set(Failed)
			x(pro.info)
			return
		} else {
			pro.info = nil
			pro.set(Stopped)
		}

		if err := pro.data.Start(ctx); err != nil {
			pro.info = fmt.Errorf("%s open fail error %v", pro.name, err)
			pro.set(Failed)
			x(pro.info)
		} else {
			pro.set(Succeed)
			pro.info = nil
		}

	default:
		if err := pro.data.Start(ctx); err != nil {
			pro.info = fmt.Errorf("%s open fail error %v", pro.name, err)
			pro.set(Failed)
			x(pro.info)
		} else {
			pro.info = nil
			pro.set(Succeed)
		}
	}

}

func (t *Task) pcall() {
	co := t.private.LState
	reader := strings.NewReader(t.config.Code)

	fn, err := co.Load(reader, "="+t.config.Name)
	if err != nil {
		t.reply.Reason = err.Error()
		return
	}

	err = co.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: t.Tree().Protect(),
	})

	if err != nil {
		t.reply.Reason = err.Error()
	}

	t.reply.Succeed = err == nil
	t.Tree().Report(t)
}

func (t *Task) do() error {
	tree := t.Tree()
	parent := tree.Context()
	ctx, cancel := context.WithTimeout(parent, t.Timeout())

	//
	t.private.Context = ctx
	t.private.Cancel = cancel

	//init lua.LState coroutine
	kit := tree.NewKit() // 功能的注入 lua 虚拟机
	t.Preload(kit)
	t.private.LState = kit.NewState(ctx, t.Key(), func(option *lua.Options) {
		option.Exdata = t
	})

	tree.Submit(t)
	return nil
}

func NewTask(t *TaskTree, config *TaskConfig) *Task {
	tas := &Task{
		config: config,
	}

	tas.private.Tree = t

	tas.reply = &Reply{
		ID:      config.ID,
		ExecID:  config.ExecID,
		Succeed: false,
		Data:    libkit.NewDataKV[string, any](),
		Reason:  "",
	}
	return tas
}
