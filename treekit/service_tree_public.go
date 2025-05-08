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

func (mt *MsTree) Debugf(format string, v ...any) {
	mt.handler.Debug.Invoke(fmt.Sprintf(format, v...))
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

func (mt *MsTree) Map() map[string]*ServiceView {
	mt.cache.mutex.RLock()
	defer mt.cache.mutex.RUnlock()
	tab := make(map[string]*ServiceView)
	sz := len(mt.cache.data)
	if sz == 0 {
		return tab
	}

	for i := 0; i < sz; i++ {
		dat := mt.cache.data[i]
		tab[dat.Key()] = dat.View()
	}
	return tab
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

func (mt *MsTree) DoDiff(v []*ServiceEntry) error {
	var diff Diff
	tab := mt.Map()

	for _, entry := range v {
		view, ok := tab[entry.Name]
		if !ok {
			diff.Updates = append(diff.Updates, entry)
			continue
		}

		delete(tab, entry.Name)
		if view.Hash == entry.Hash {
			diff.Nothing = append(diff.Nothing, view)
			continue
		}

		diff.Updates = append(diff.Updates, entry)
	}

	for _, view := range tab {
		diff.Removes = append(diff.Removes, &ServiceEntry{
			Name:  view.Name,
			ID:    view.ID,
			MTime: view.MTime,
		})
	}

	linked := func(filter func(...string) bool, ss []*ServiceEntry, todo func(link string)) {
		for _, entry := range ss {
			if filter(entry.Name) {
				todo(entry.Name)
			}
		}
	}

	for _, mod := range diff.Nothing {
		ms, ok := mt.find(mod.Name)
		if !ok {
			ms.NoError("can view %s service but not found service", mod.Name)
			continue
		}

		if len(ms.config.Source) == 0 {
			continue
		}

		linked(ms.BeLink, diff.Updates, func(link string) {
			if !ms.HasSource() {
				mt.Errorf("want update %s service cause linked %s but not found source", ms.Key(), link)
				return
			}
			diff.Updates = append(diff.Updates, ms.Again())
			mt.Debugf("update %s service cause linked %s", ms.Key(), link)
		})

		linked(ms.BeLink, diff.Removes, func(link string) {
			if !ms.HasSource() {
				mt.Errorf("want remove %s service cause linked %s but not found source", ms.Key(), link)
				return
			}
			diff.Updates = append(diff.Updates, ms.Again())
			mt.Debugf("update %s service cause linked %s", ms.Key(), link)
		})
	}

	return mt.update(diff)
}

func (mt *MsTree) update(d Diff) error {
	if d.NotChange() {
		return nil
	}

	//diff remove task by ids
	mt.Remove(func(ms *MicroService) bool {
		for _, entry := range d.Removes {
			if entry.Name == ms.Key() {
				return true
			}
		}
		return false
	})

	errs := errkit.New()
	for _, entry := range d.Updates {
		if e := mt.Register(entry.Name, entry.Chunk, func(c *MicoServiceConfig) {
			c.ID = entry.ID
			c.Hash = entry.Hash
			c.Dialect = entry.Dialect
			c.MTime = entry.MTime
		}); e != nil {
			errs.Try(entry.Name, e)
		}
	}

	if e := errs.Wrap(); e != nil {
		mt.Errorf(e.Error())
	}
	mt.Wakeup()
	return mt.UnwrapErr()
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

func (mt *MsTree) Load(v ...Script) error {
	errs := errkit.New()
	s, err := Load(v...)
	if err != nil {
		errs.Try("load", err)
	}

	err = mt.DoDiff(s)
	if err != nil {
		errs.Try("diff", err)
	}

	return errs.Wrap()
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

func (mt *MsTree) DoString(v *LuaText) error {
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

func (mt *MsTree) Reload(filter func(name string) bool) error {
	sz := len(mt.cache.data)
	if sz == 0 {
		return nil
	}
	mt.cache.mutex.RLock()
	defer mt.cache.mutex.RUnlock()

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

func (mt *MsTree) Diff(info []ServiceDiffInfo) error {
	return nil
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
		err := mt.SafeWakeup(tas)
		errs.Try(tas.Key(), err)
	}

	mt.private.error = errs.Wrap()
}
