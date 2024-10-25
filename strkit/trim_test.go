package strkit

import (
	"testing"
)

func TestGenN(t *testing.T) {
	s := &Trim{
		Text:   "今天 2024",
		Handle: []TrimFunc{NewTrimN()},
	}

	t.Logf("%s", s.ToMask())
}

//func TestGenMapping(t *testing.T) {
//	ac, err := NewAcFile("name.txt")
//	if err != nil {
//		t.Log(err)
//		return
//	}
//
//	text := "你好项建国告诉黄运发今天天气怎么样?快要很好就是里面的东西"
//	t.Logf("%s", text)
//
//	r := ac.Replace([]byte(text), []byte("EDD"))
//	t.Log(string(r))
//}
//
//func TestNewTrimRegex(t *testing.T) {
//
//	a := false
//	b := true
//
//	t.Logf("%v", a && b)
//
//}
