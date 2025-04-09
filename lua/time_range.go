package lua

import (
	"fmt"
	"strings"
	"time"
)

func (rt *RangeTime) layoutTime(s string, e string) error {
	start, err := time.Parse("15:04", s)
	if err != nil {
		return err
	}
	end, err := time.Parse("15:04", e)
	if err != nil {
		return err
	}

	rt.StartFunc = func(v Time) Time {
		t := time.Time(v)
		r := time.Date(t.Year(), t.Month(), t.Day(), start.Hour(), start.Minute(), start.Second(), 0, t.Location())
		return Time(r)
	}

	rt.EndFunc = func(v Time) Time {
		t := time.Time(v)
		et := time.Date(t.Year(), t.Month(), t.Day(), end.Hour(), end.Minute(), end.Second(), 0, t.Location())
		return Time(et)
	}

	return nil
}

func (rt *RangeTime) parse(s string) error {
	switch s {
	case "default":
		rt.StaticTimeSlot = true
		return rt.layoutTime("9:30", "15:00")
	case "weekend":
		rt.Weekend = true
		return nil

	default:
		rt.StaticTimeSlot = true
		parts := strings.Split(s, "-")
		if len(parts) != 2 {
			return fmt.Errorf("static time solt must be xx:xx-xx:xx got %s", s)
		}
		return rt.layoutTime(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}
}

// 如果s是default，则时间判断为9:30-15:00
// 如果s是weekend, 则时间判断为周六00:00-周日23:59
// 如果s是xx:xx-xx:xx，则另外判断
// 除weekend外，layout均为15:04:05，s中自动补齐
func (r *RangeTimes) compile() error {
	sz := len(r.Texts)
	if sz == 0 {
		return nil
	}

	r.Times = make([]RangeTime, 0, sz)

	for i := 0; i < sz; i++ {
		rt := RangeTime{}
		text := r.Texts[i]
		err := rt.parse(text)
		if err != nil {
			return err
		}
		r.Times = append(r.Times, rt)
	}

	return nil
}

func NewRange(times []string) *RangeTimes {
	r := &RangeTimes{
		Texts: times,
	}
	return r
}
