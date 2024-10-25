package strkit

import (
	"math/rand"
	"testing"
)

func TestAc(t *testing.T) {
	dict := []string{"你好", "VeLa"}
	ac := NewAc(dict, false)
	ac.Build()

	data := "yesvelanihaovelaabc sorry"
	indexes := ac.Match(data)
	for _, term := range indexes {
		t.Logf("%s\n", data[term.From:term.To])
	}

}

func Test1(t *testing.T) {
	ac := NewAcMatcher(false)

	dictionary := []string{"she", "he", "say", "shr", "her"}
	ac.From(dictionary)

	expected := []*Term{
		{Index: 0, To: 5},
		{Index: 1, To: 5},
		{Index: 4, To: 6},
	}

	s := "yasherhs"
	ret := ac.Match(s)
	if len(expected) != len(ret) {
		t.Fatal()
	}
	for i, _ := range ret {
		if ret[i].Index != expected[i].Index || ret[i].To != expected[i].To {
			t.Fatal()
		}

		original := dictionary[ret[i].Index]
		matched := s[ret[i].To-len(original) : ret[i].To]
		if original != matched {
			t.Fatal()
		}
	}
}

func Test2(t *testing.T) {
	ac := NewAcMatcher(false)

	dictionary := []string{"中国人民", "国人", "中国人", "hello世界", "hello"}
	ac.From(dictionary)

	if len(ac.Match("中国人")) != 2 {
		t.Fatal()
	}
	if len(ac.Match("世界")) != 0 {
		t.Fatal()
	}

	s := "hello世界"
	ret := ac.Match(s)
	if len(ret) != 2 {
		t.Fatal()
	}

	for i, _ := range ret {
		original := dictionary[ret[i].Index]
		matched := s[ret[i].To-len(original) : ret[i].To]
		if original != matched {
			t.Fatal()
		}
	}
}

func Benchmark(b *testing.B) {
	ac := NewAcMatcher(false)

	dictionary := make([]string, 0)
	for i := 0; i < 200000; i++ {
		dictionary = append(dictionary, randWord(2, 6))
	}
	ac.From(dictionary)

	for i := 0; i < b.N; i++ {
		ac.Match(randWord(5000, 10000))
	}
}

func randWord(m, n int) string {
	num := rand.Intn(n-m) + m
	var s string
	var a rune = 'a'

	for i := 0; i < num; i++ {
		c := a + rune(rand.Intn(26))
		s = s + string(c)
	}

	return s
}
