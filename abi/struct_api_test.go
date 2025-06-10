package abi

import (
	"fmt"
	"testing"
)

type A struct {
	age   int32
	name  [12]byte
	score float32
}

func (a *A) print() {
	fmt.Printf("age:%d score:%f name:%s\n", a.age, a.score, BytesToCleanString(a.name[:]))
}

func TestABI(t *testing.T) {
	bu := NewStructBuilder(false)

	_ = bu.Define("age:int32")
	_ = bu.Define("name:text(12)")
	_ = bu.Define("score:float32")

	s, _ := bu.Finalize()

	_ = s.SetInt32("age", 12)
	_ = s.SetFloat32("score", 95.5)
	_ = s.SetText("name", "hello world")

	t.Logf("%#v", s.ToMap())

	cdata, ok := Cast[A](s)
	if !ok {
		t.Error("cast error")
	}

	cdata.print()
}
