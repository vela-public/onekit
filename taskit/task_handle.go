package taskit

import (
	"fmt"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/todo"
	"strings"
)

var (
	ErrTaskKeyEmpty = fmt.Errorf("not found task key")
)

func (t *task) link(name string) {
	if !libkit.In(t.service.Link, name) {
		t.service.Link = append(t.service.Link, name)
	}
}

func (t *task) NoError(format string, v ...any) {
	t.root.Errorf(format, v...)
}

func (t *task) way() string {
	return todo.IF[string](t.config.Path == "", "本地", "远程")
}

func (t *task) View() *TaskView {
	tv := &TaskView{
		ID:      t.config.ID,
		Name:    t.Key(),
		Hash:    t.config.Hash,
		Link:    strings.Join(t.service.Link, ","),
		From:    t.way(),
		Status:  t.private.Flag.String(),
		Uptime:  t.private.Uptime,
		Dialect: t.config.Dialect,
	}

	if e := t.UnwrapErr(); e != nil {
		tv.Failed = true
		tv.Cause = e.Error()
	} else {
		tv.Failed = false
	}

	for _, srv := range t.service.store {
		r := &Runner{
			Name:     srv.name,
			Type:     srv.data.TypeOf(),
			Status:   srv.Status(),
			CodeVM:   t.Key(),
			Private:  srv.private,
			Metadata: srv.data.Metadata(),
		}

		if srv.info != nil {
			r.Cause = srv.info.Error()
		}
		tv.Runners = append(tv.Runners, r)
	}
	return tv
}
