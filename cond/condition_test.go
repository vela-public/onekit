package cond

import (
	"strconv"
	"testing"
)

type Event struct {
	Addr  string
	Type  string
	Value int
}

func (ev *Event) Field(key string) string {
	switch key {
	case "type":
		return ev.Type
	case "value":
		return strconv.Itoa(ev.Value)
	case "addr":
		return ev.Addr
	}

	return ""
}

func TestExp(t *testing.T) {
	cnd := NewText("value ~ (.*)")
	ev := &Event{
		Type:  "typeof",
		Value: 456,
	}

	pay := func(id int, ret string) {
		t.Logf("%d %v", id, ret)
	}

	t.Log(cnd.Match(ev, Payload(pay)))
}

func TestUnary(t *testing.T) {
	cnd := NewText("type = typeof")
	ev := &Event{
		Type:  "typeof",
		Value: 456,
		Addr:  "a",
	}

	t.Log(cnd.Match(ev))
}

func TestString(t *testing.T) {

	raw := "12-345-67.raw"

	pbc := String(raw)
	ext := pbc("[:6]")

	t.Log(ext)

}
func TestRegex(t *testing.T) {
	val := "10.10.239.11"
	cnd := NewText("[0,13] ~ \\.(.*)\\.(.*)\\.(.*)")

	pay := func(id int, ret string) {
		t.Logf("%d %v", id, ret)
	}

	t.Log(cnd.Match(val, Payload(pay)))
}
