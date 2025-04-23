package treekit

import (
	"encoding/json"
	"fmt"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/libkit"
	"os"
)

func (mt *MsTree) UnwrapErr() error {
	return mt.private.error
}

func (mt *MsTree) Errorf(format string, v ...any) {
	mt.handler.Error.Invoke(fmt.Errorf(format, v...))
}

func (mt *MsTree) OnError(err error) {
	mt.handler.Error.Invoke(err)
}

func (mt *MsTree) OnCreate(srv *Process) {
	mt.handler.Create.Invoke(srv)
}

func (mt *MsTree) OnWakeup(srv *Process) {
	mt.handler.Wakeup.Invoke(srv)
}

func (mt *MsTree) View() *TreeView {
	mt.cache.mutex.RLock()
	defer mt.cache.mutex.RUnlock()
	tv := new(TreeView)
	sz := len(mt.cache.data)
	if sz == 0 {
		return tv
	}
	for i := 0; i < sz; i++ {
		tas := mt.cache.data[i]
		tv.Services = append(tv.Services, tas.View())
	}
	return tv
}

func (mt *MsTree) Doc() []byte {
	v := mt.View()
	text, _ := json.Marshal(v)
	return text
}

func (mt *MsTree) create(config *MicoServiceConfig) (*MicroService, error) {
	if config.Key == "" {
		return nil, ErrServiceKeyEmpty
	}

	tas, ok := mt.find(config.Key)
	if !ok {
		tas = &MicroService{
			root:   mt,
			config: config,
		}
		tas.build()
		mt.push(tas)
		return tas, nil
	}
	tas.update(config)
	return tas, tas.UnwrapErr()
}

func (mt *MsTree) DoServiceFile(key string, path string) error {
	cfg, err := NewFile(key, path)
	if err != nil {
		return err
	}
	tas, err := mt.create(cfg)
	if err != nil {
		return err
	}
	return tas.wakeup()
}
func (mt *MsTree) DoString(v LuaText) error {
	cfg, err := NewText(v.Name, v.Text, v.MTime.Unix())
	if err != nil {
		return err
	}
	tas, err := mt.create(cfg)
	if err != nil {
		return err
	}
	return tas.wakeup()
}

func (mt *MsTree) ReloadText(filter func(name string) bool, needle func(name string) LuaText) error {
	sz := len(mt.cache.data)
	if sz == 0 {
		return nil
	}

	errs := &errkit.JoinError{}
	for i := 0; i < sz; i++ {
		tas := mt.cache.data[i]
		if filter != nil && !filter(tas.Key()) {
			continue
		}

		v := needle(tas.Key())
		cnf := tas.config
		if cnf.MTime == 0 {
			continue
		}

		mtime := v.MTime.Unix()
		if mtime != cnf.MTime {
			errs.Try(tas.Key(), mt.DoString(v))
		}
	}
	return errs.Wrap()

}

func (mt *MsTree) ReloadFile(filter func(name string) bool) error {
	sz := len(mt.cache.data)
	if sz == 0 {
		return nil
	}

	errs := &errkit.JoinError{}
	for i := 0; i < sz; i++ {
		tas := mt.cache.data[i]
		if filter != nil && !filter(tas.Key()) {
			continue
		}

		st, err := os.Stat(tas.config.Path)
		if err != nil {
			errs.Try(tas.Key(), fmt.Errorf("file not found %s", tas.config.Path))
			continue
		}

		cnf := tas.config
		if cnf.MTime == 0 {
			continue
		}

		if st.ModTime().Unix() != cnf.MTime {
			errs.Try(tas.Key(), mt.DoServiceFile(tas.config.Key, tas.config.Path))
		}
	}
	return errs.Wrap()
}

func (mt *MsTree) Register(key string, body []byte, options ...func(*MicoServiceConfig)) error {
	cfg, err := NewConfig(key, func(c *MicoServiceConfig) {
		c.Source = body
		for _, option := range options {
			option(c)
		}
	})

	if err != nil {
		return err
	}
	tas, err := mt.create(cfg)
	if err != nil {
		return err
	}

	return tas.UnwrapErr()
}

func (mt *MsTree) SafeWakeup(tas *MicroService) (err error) {
	if mt.Protect() {
		defer func() {
			if e := recover(); e != nil {
				tas.put(Panic)
				tas.private.Stack = fmt.Sprintf("%v\n%s", e, libkit.StackTrace[string](1024*1024, false))
			}
		}()
	}
	err = tas.wakeup()
	return
}

func (mt *MsTree) report() {
	mt.handler.Report.Invoke(mt)
}

func (mt *MsTree) Wakeup() {
	mt.cache.mutex.RLock()
	defer mt.cache.mutex.RUnlock()
	sz := len(mt.cache.data)
	if sz == 0 {
		return
	}

	errs := &errkit.JoinError{}
	for i := 0; i < sz; i++ {
		tas := mt.cache.data[i]
		errs.Try(tas.Key(), mt.SafeWakeup(tas))
	}

	mt.private.error = errs.Wrap()
}
