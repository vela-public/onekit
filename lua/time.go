package lua

import (
	"time"
)

const (
	TimeLibName = "time"
)

/*
	local t = time.now()
	local a = time.now()
	local t1 = time("2022-01-01 12:00:00")
	local t2 = time(1002340)
	local t3 = time()

*/

type Layout struct {
	Format string
}

func (l *Layout) Parse(L *LState) Time {
	v := L.Get(1)
	switch v.Type() {
	case LTNumber:
		t := time.Unix(int64(v.(LNumber)), 0)
		return Time(t)
	case LTInt64:
		t := time.Unix(int64(v.(LInt64)), 0)
		return Time(t)
	case LTString:
		t, err := time.Parse(l.Format, v.String())
		if err != nil {
			return Time(time.Unix(0, 0))
		}
		return Time(t)
	case LTObject:
		t, ok := v.(Time)
		if !ok {
			return Time(time.Unix(0, 0))
		}
		return t
	default:
		return Time(time.Unix(0, 0))
	}
}

func (l *Layout) AssertFunction() (*LFunction, bool) {
	return NewFunction(func(L *LState) int {
		t := l.Parse(L)
		L.Push(t)
		return 1
	}), true
}

func NewLayoutL(L *LState) int {

	l := L.IsString(1)
	if l == "" {
		l = "2006-01-02 15:04:05"
	}
	L.Push(NewGeneric(&Layout{
		Format: l,
	}))
	return 1
}

type Time time.Time

func (t Time) String() string                     { return time.Time(t).Format("2006-01-02 15:04:05") }
func (t Time) Type() LValueType                   { return LTObject }
func (t Time) AssertFloat64() (float64, bool)     { return float64(time.Time(t).Unix()), true }
func (t Time) AssertString() (string, bool)       { return "", false }
func (t Time) AssertFunction() (*LFunction, bool) { return nil, false }
func (t Time) Hijack(fsm *CallFrameFSM) bool      { return false }
func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}

func (t Time) Index(L *LState, key string) LValue {
	switch key {
	case "format":
		return NewFunction(func(L *LState) int {
			layout := L.IsString(1)
			if layout == "" {
				layout = "2006-01-02 15:04:05"
			}
			L.Push(S2L(time.Time(t).Format(layout)))
			return 1
		})

	case "unix":
		return LInt64(time.Time(t).Unix())
	case "today":
		return S2L(time.Time(t).Format("2006-01-02"))
	case "year":
		return LInt(time.Time(t).Year())
	case "month":
		return LInt(time.Time(t).Month())
	case "weekday":
		return LInt(time.Time(t).Weekday())
	case "day":
		return LInt(time.Time(t).Day())
	case "hour":
		return LInt(time.Time(t).Hour())
	case "minute":
		return LInt(time.Time(t).Minute())
	case "second":
		return LInt(time.Time(t).Second())
	case "scope":
		return NewFunction(t.ScopeL)
	default:
		return LNil
	}
}

// time.scope("2025-0101", "2025-0102" , "2006-01-02")
func (t Time) ScopeL(L *LState) int {

	layout := TimeLayout(L, 3)
	s := TimeExL(L, 1, layout)
	e := TimeExL(L, 2, layout)

	if n := t.Unix(); n != 0 && s.Unix() <= n && n <= e.Unix() {
		L.Push(LTrue)
		return 1
	}

	L.Push(LFalse)
	return 1
}

func TimeLayout(L *LState, idx int) string {
	d := L.IsString(idx)
	if d == "" {
		d = "2006-01-02 15:04:05"
	}
	return d
}

func TimeExL(L *LState, idx int, layout string) Time {
	v := L.Get(idx)
	switch v.Type() {
	case LTNumber:
		t := time.Unix(int64(v.(LNumber)), 0)
		return Time(t)
	case LTInt64:
		t := time.Unix(int64(v.(LInt64)), 0)
		return Time(t)
	case LTString:
		t, err := time.Parse(layout, v.String())
		if err != nil {
			return Time(time.Unix(0, 0))
		}
		return Time(t)
	case LTObject:
		t, ok := v.(Time)
		if !ok {
			return Time(time.Unix(0, 0))
		}
		return t
	default:
		return Time(time.Unix(0, 0))
	}

}

func NewTimeL(L *LState) int {
	v := L.Get(1)
	layout := TimeLayout(L, 2)

	switch v.Type() {
	case LTNumber:
		t := time.Unix(int64(v.(LNumber)), 0)
		L.Push(Time(t))
		return 1
	case LTInt64:
		t := time.Unix(int64(v.(LInt64)), 0)
		L.Push(Time(t))
		return 1
	case LTString:
		t, err := time.Parse(layout, v.String())
		if err != nil {
			L.Push(Time(time.Unix(0, 0)))
			L.Push(S2L(err.Error()))
			return 2
		}
		L.Push(Time(t))
		return 1
	case LTObject:
		t, ok := v.(Time)
		if !ok {
			L.Push(Time(time.Unix(0, 0)))
			L.Push(S2L("not time object fail"))
			return 2
		}
		L.Push(t)
		return 1
	default:
		L.Push(Time(time.Now()))
		return 1
	}
}

func TimeNowL(L *LState) int {
	L.Push(Time(time.Now()))
	return 1
}

func TimeUnixL(L *LState) int {
	L.Push(LInt64(time.Now().Unix()))
	return 1
}

func TimeDayL(L *LState) int {
	n := L.IsInt(1)
	t := time.Now().AddDate(0, 0, n)
	L.Push(Time(t))
	return 1
}

func NewTimeSleepL(L *LState) int {
	n := L.IsInt(1)
	if n <= 0 {
		return 0
	}

	time.Sleep(time.Duration(n) * time.Millisecond)
	return 0
}

func NewTimeRangeL(L *LState) int {
	times := Unpack[string](L)
	tr := NewRange(times)
	if e := tr.compile(); e != nil {
		L.RaiseError("%v", e)
		return 0
	}
	L.Push(tr)
	return 1
}

func TimeIndexL(L *LState, key string) LValue {
	switch key {
	case "now":
		return NewFunction(TimeNowL)
	case "unix":
		return NewFunction(TimeUnixL)
	case "today":
		return S2L(time.Now().Format("2006-01-02"))
	case "tomorrow":
		return S2L(time.Now().AddDate(0, 0, 1).Format("2006-01-02"))
	case "day":
		return NewFunction(TimeDayL)
	case "layout":
		return NewFunction(NewLayoutL)
	case "range":
		return NewFunction(NewTimeRangeL)
	case "sleep":
		return NewFunction(NewTimeSleepL)
	}
	return LNil
}

func OpenTimeLib(L *LState) int {
	mod := NewExport("lua.time.export", WithFunc(NewTimeL), WithIndex(TimeIndexL))
	L.SetGlobal("time", mod)
	L.Push(mod)
	return 1
}
