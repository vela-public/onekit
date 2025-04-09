package lua

import (
	"strings"
	"time"
)

type RangeTime struct {
	// 如果time range为固定时间段，StaticTimeSlot置为true
	StaticTimeSlot bool
	StartFunc      func(Time) Time
	EndFunc        func(Time) Time

	// 如果time range为周末，Weekend置为true
	Weekend bool
}

func (rt *RangeTime) Match(v time.Time) bool {
	if rt.Weekend {
		if week := v.Weekday(); week == time.Saturday || week == time.Sunday {
			return true
		}
	}

	if !rt.StaticTimeSlot {
		return false
	}

	st := rt.StartFunc(Time(v))
	et := rt.EndFunc(Time(v))
	if v.After(time.Time(st)) && v.Before(time.Time(et)) {
		return true
	}

	return false
}

type RangeTimes struct {
	Texts []string
	Times []RangeTime
}

func (r *RangeTimes) Match(v Time) bool {
	sz := len(r.Times)
	if sz == 0 {
		return false
	}

	for i := 0; i < sz; i++ {
		rt := r.Times[i]
		if rt.Match(time.Time(v)) {
			return true
		}
	}
	return false
}

func (r *RangeTimes) NowL(L *LState) int {
	now := Time(time.Now())
	m := r.Match(now)
	L.Push(LBool(m))
	return 1
}
func (r *RangeTimes) TimeL(L *LState) int {
	tv := TimeExL(L, 1, TimeLayout(L, 2))
	tm := time.Time(tv)
	if tm.IsZero() {
		L.RaiseError("not time #1")
		return 0
	}
	ma := r.Match(tv)
	L.Push(LBool(ma))
	return 1
}

func (r *RangeTimes) String() string                     { return strings.Join(r.Texts, ",") }
func (r *RangeTimes) Type() LValueType                   { return LTObject }
func (r *RangeTimes) AssertFloat64() (float64, bool)     { return 0, false }
func (r *RangeTimes) AssertString() (string, bool)       { return "", false }
func (r *RangeTimes) AssertFunction() (*LFunction, bool) { return NewFunction(r.TimeL), true }
func (r *RangeTimes) Hijack(fsm *CallFrameFSM) bool      { return false }

func (r *RangeTimes) Index(L *LState, key string) LValue {
	switch key {
	case "now":
		return NewFunction(r.NowL)
	case "time":
		return NewFunction(r.TimeL)
	}
	return LNil
}
