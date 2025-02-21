package treekit

import (
	"fmt"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/todo"
	"strings"
)

var (
	ErrServiceKeyEmpty = fmt.Errorf("not found task key")
)

func (ms *MicroService) link(name string) {
	if !libkit.In(ms.processes.Link, name) {
		ms.processes.Link = append(ms.processes.Link, name)
	}
}

func (ms *MicroService) NoError(format string, v ...any) {
	ms.root.Errorf(format, v...)
}

func (ms *MicroService) way() string {
	return todo.IF[string](ms.config.Path == "", "本地", "远程")
}

func (ms *MicroService) View() *ServiceView {
	tv := &ServiceView{
		ID:      ms.config.ID,
		Name:    ms.Key(),
		Hash:    ms.config.Hash,
		Link:    strings.Join(ms.processes.Link, ","),
		From:    ms.way(),
		Status:  ms.private.Flag.String(),
		Uptime:  ms.private.Uptime,
		Dialect: ms.config.Dialect,
	}

	if e := ms.UnwrapErr(); e != nil {
		tv.Failed = true
		tv.Cause = e.Error()
	} else {
		tv.Failed = false
	}

	sz := len(ms.processes.data)
	if sz == 0 && ms.has(Running) {
		tv.Status = Empty.String()
		return tv
	}

	for _, pro := range ms.processes.data {
		r := &Runner{
			Name:     pro.name,
			Type:     pro.data.TypeOf(),
			Status:   pro.Status(),
			CodeVM:   pro.From(),
			Private:  pro.private,
			Metadata: pro.data.Metadata(),
		}

		if pro.info != nil {
			r.Cause = pro.info.Error()
		}
		tv.Runners = append(tv.Runners, r)
	}

	return tv
}
