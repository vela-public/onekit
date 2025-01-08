package taskit

import (
	"encoding/json"
	"fmt"
)

func (t *Tree) UnwrapErr() error {
	return t.private.error
}

func (t *Tree) Errorf(format string, v ...any) {
	t.handler.Error.Invoke(fmt.Errorf(format, v...))
}

func (t *Tree) OnError(err error) {
	t.handler.Error.Invoke(err)
}

func (t *Tree) OnCreate(srv *Service) {
	t.handler.Create.Invoke(srv)
}

func (t *Tree) OnWakeup(srv *Service) {
	t.handler.Wakeup.Invoke(srv)
}

func (t *Tree) View() *TreeView {
	t.cache.mutex.RLock()
	defer t.cache.mutex.RUnlock()
	tv := new(TreeView)
	sz := len(t.cache.data)
	if sz == 0 {
		return tv
	}
	for i := 0; i < sz; i++ {
		tas := t.cache.data[i]
		tv.Tasks = append(tv.Tasks, tas.View())
	}
	return tv
}

func (t *Tree) Doc() []byte {
	v := t.View()
	text, _ := json.Marshal(v)
	return text
}
