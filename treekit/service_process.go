package treekit

import (
	"fmt"
	"github.com/vela-public/onekit/libkit"
	"runtime"
)

type ProcessType interface {
	Name() string
	Start(*Env) error
	Close() error
	Metadata() libkit.DataKV[string, any]
}

type Process struct {
	name    string
	flag    ErrNo //0: init 1: running 2: stop 3: error
	info    error
	from    string
	typeof  string
	private bool
	data    ProcessType
}

func (pro *Process) Name() string {
	return pro.data.Name()
}

func (pro *Process) Nil() bool {
	return pro.data == nil
}

func (pro *Process) From() string {
	return pro.from
}

func (pro *Process) Set(s ProcessType) {
	pro.data = s
}

func (pro *Process) UnwrapErr() error {
	return pro.info
}

func (pro *Process) Unpack() any {
	return pro.data
}

func (pro *Process) Errorf(format string, v ...any) {
	pro.info = fmt.Errorf(format, v...)
}

func (pro *Process) Close() error {
	if pro.has(Stopped) {
		return nil
	}

	defer func() {
		if e := recover(); e != nil {
			pro.set(Stopped | Failed)
			buf := make([]byte, 1024*1024)
			n := runtime.Stack(buf, false)
			pro.info = fmt.Errorf("%v\n%s", e, buf[:n])
		}
	}()

	err := pro.data.Close()
	if err != nil {
		pro.set(Stopped | Failed)
		pro.info = fmt.Errorf("stop fail:%s", err.Error())
	} else {
		pro.set(Stopped)
		pro.info = fmt.Errorf("stopped")
	}

	return err
}

func (pro *Process) Status() string {
	return pro.flag.String()
}

func (pro *Process) is(op ErrNo) bool {
	return pro.flag&op == op
}

func (pro *Process) put(op ErrNo) {
	pro.flag = pro.flag | op
}

func (pro *Process) has(op ErrNo) bool {
	return pro.flag&op == op
}

func (pro *Process) set(op ErrNo) {
	pro.flag = op
}

func (pro *Process) Reload() {
	pro.flag = Reload
}

func (pro *Process) Update(fn func(p ProcessType)) {
	if fn == nil {
		return
	}
	fn(pro.data)
}

func (pro *Process) Call(fn func(p ProcessType)) {
	if fn == nil {
		return
	}

	fn(pro.data)
}
