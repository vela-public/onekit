package abi

import (
	"fmt"
	"testing"
)

type A struct {
	age   int32
	name  [12]byte
	score int32
}

func TestABI(t *testing.T) {
	bu := NewStructBuilder(false)

	_ = bu.Define("age:int32")
	_ = bu.Define("name:text(12)")
	_ = bu.Define("score:int32")

	s, _ := bu.Finalize()

	_ = s.SetInt32("age", 12)
	_ = s.SetInt32("score", 100)
	_ = s.SetText("name", "hello world")

	t.Logf("%#v", s.ToMap())

	cdata, ok := Cast[A](s)
	if !ok {
		t.Error("cast error")
	}

	fmt.Printf("age:%d score:%d name:%s\n", cdata.age, cdata.score, BytesToCleanString(cdata.name[:]))

}
