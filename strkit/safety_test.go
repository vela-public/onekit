package strkit

import "testing"

func TestDemo(t *testing.T) {
	r := []string{"123", "123"}
	for i := 0; i < len(r); i++ {
		var path []byte
		path = append(path, []byte(r[i])...)
		println(string(path))
	}

}

func TestSafety(t *testing.T) {
	/*
		S[2,3]N3.txt

	*/
	demo := []string{
		"select 1,2,3,4,5,6#",
		//"/abc/a.txt",
		//"/aa011/你好000001.txt?path=../../etc/passwd",
		"http://www.eastmoney.com/api/user?page=/etc/passwd&name={{/bin/bash -i}}",
	}

	s := NewSafety()
	s.Holds("/.&=")
	s.Short()
	s.SQLi()
	s.Xss()
	s.UnescapeUri()
	s.Bad("../", "/bin/bash", "/etc/passwd")
	s.Keyword("https://", "http://")
	s.Build()
	for i := 0; i < len(demo); i++ {
		fsm := s.Do(demo[i])
		t.Logf("norm:%s simple:%s detail:%s ext:%s bad: %v", fsm.Norm(), fsm.Simplicity(), fsm.Detail(), fsm.Ext, fsm.Bad)
	}
}
