package taskit

import (
	"fmt"
	"github.com/vela-public/onekit/libkit"
	"runtime"
)

type ServiceType interface {
	Name() string
	Start() error
	Close() error
	TypeOf() string
	Metadata() libkit.DataKV[string, any]
}

type Service struct {
	name    string
	flag    ErrNo //0: init 1: running 2: stop 3: error
	info    error
	from    string
	private bool
	data    ServiceType
}

func (srv *Service) Nil() bool {
	return srv.data == nil
}

func (srv *Service) From() string {
	return srv.from
}

func (srv *Service) Set(s ServiceType) {
	srv.data = s
}

func (srv *Service) UnwrapErr() error {
	return srv.info
}

func (srv *Service) UnwrapData() any {
	return srv.data
}

func (srv *Service) Errorf(format string, v ...any) {
	srv.info = fmt.Errorf(format, v...)
}

func (srv *Service) Close() error {
	if srv.has(Stopped) {
		return nil
	}

	defer func() {
		if e := recover(); e != nil {
			srv.set(Stopped | Failed)
			buf := make([]byte, 1024*1024)
			n := runtime.Stack(buf, false)
			srv.info = fmt.Errorf("%v\n%s", e, buf[:n])
		}
	}()

	err := srv.data.Close()
	if err != nil {
		srv.set(Stopped | Failed)
		srv.info = err
	}
	return err
}

func (srv *Service) Status() string {
	return srv.flag.String()
}

func (srv *Service) is(op ErrNo) bool {
	return srv.flag&op == op
}

func (srv *Service) put(op ErrNo) {
	srv.flag = srv.flag | op
}

func (srv *Service) has(op ErrNo) bool {
	return srv.flag&op == op
}

func (srv *Service) set(op ErrNo) {
	srv.flag = op
}

func (srv *Service) Reload() {
	srv.flag = Reload
}

func (srv *Service) Call(fn func(srv ServiceType)) {
	if fn == nil {
		return
	}

	fn(srv.data)
}
