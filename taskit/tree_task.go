package taskit

import (
	"fmt"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/libkit"
)

func (t *Tree) create(config *Config) (*task, error) {
	if config.Key == "" {
		return nil, ErrTaskKeyEmpty
	}

	tas, ok := t.find(config.Key)
	if !ok {
		tas = &task{
			root:   t,
			config: config,
		}
		tas.build()
		t.push(tas)
		return tas, nil
	}
	tas.update(config)
	return tas, tas.UnwrapErr()
}

func (t *Tree) DoTaskFile(key string, path string) error {
	cfg, err := NewFile(key, path)
	if err != nil {
		return err
	}
	tas, err := t.create(cfg)
	if err != nil {
		return err
	}
	return tas.wakeup()
}

func (t *Tree) Register(key string, body []byte, options ...func(*Config)) error {
	cfg, err := NewConfig(key, func(c *Config) {
		c.Source = body
		for _, option := range options {
			option(c)
		}
	})

	if err != nil {
		return err
	}
	tas, err := t.create(cfg)
	if err != nil {
		return err
	}

	return tas.UnwrapErr()
}

func (t *Tree) SafeWakeup(tas *task) (err error) {
	if t.Protect() {
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

func (t *Tree) report() {
	t.handler.Report.Invoke(t)
}

func (t *Tree) Wakeup() {
	t.cache.mutex.RLock()
	defer t.cache.mutex.RUnlock()
	sz := len(t.cache.data)
	if sz == 0 {
		return
	}

	errs := &errkit.JoinError{}
	for i := 0; i < sz; i++ {
		tas := t.cache.data[i]
		errs.Try(tas.Key(), t.SafeWakeup(tas))
	}

	t.private.error = errs.Wrap()
}
