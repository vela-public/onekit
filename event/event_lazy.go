package event

import "github.com/vela-public/onekit/layer"

type LazyEvent struct {
	xEnv   layer.Environment
	option []func(*Event)
}

func Lazy(option ...func(*Event)) *LazyEvent {
	return &LazyEvent{
		xEnv:   layer.LazyEnv(),
		option: option,
	}
}

func (le *LazyEvent) apply(ev *Event) {
	sz := len(le.option)
	for i := 0; i < sz; i++ {
		le.option[i](ev)
	}
}

func (le *LazyEvent) Error(format string, v ...interface{}) *Event {
	ev := Error(le.xEnv, format, v...)
	le.apply(ev)
	return ev
}

func (le *LazyEvent) Debug(format string, v ...interface{}) *Event {
	ev := Debug(le.xEnv, format, v...)
	le.apply(ev)
	return ev
}

func (le *LazyEvent) Trace(format string, v ...interface{}) *Event {
	ev := Trace(le.xEnv, format, v...)
	le.apply(ev)
	return ev
}

func (le *LazyEvent) Create(typeof string) *Event {
	ev := NewEvent(le.xEnv, typeof)
	le.apply(ev)
	return ev
}
