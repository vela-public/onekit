package strkit

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	N0 = '0'
	N9 = '9'
	NA = 'a'
	NZ = 'z'
)

func NewTrimN() TrimFunc {
	return func(str string) string {
		u := []rune(str)
		n := len(u)
		for i := 0; i < n; i++ {
			if u[i] >= N0 && u[i] <= N9 {
				u[i] = 'N'
			}
		}
		return string(u)
	}
}

func NewTrimGraphic(flag bool) TrimFunc {
	return func(str string) string {
		return strings.TrimFunc(str, func(r rune) bool {
			return !unicode.IsGraphic(r)
		})
	}
}

func NewTrimSpace() TrimFunc {
	return func(s string) string {
		return strings.TrimFunc(s, func(r rune) bool {
			return unicode.IsSpace(r)
		})
	}
}

func NewTrimAcFile(filename string, ch string) TrimFunc {
	return func(s string) string {
		//ac, err := NewAcFile(filename)
		//if err != nil {
		//	return s
		//}
		//return cast.B2S(ac.Replace(cast.S2B(s), cast.S2B(ch)))
		return ""
	}
}

func NewTrimRegex(regex string, ch string) TrimFunc {
	return func(s string) string {
		re, err := regexp.Compile(regex)
		if err != nil {
			return s
		}

		return re.ReplaceAllString(s, ch)
	}
}

func NewTrimDate(f string, m string) TrimFunc {
	return func(str string) string {
		return strings.ReplaceAll(str, f, m)
	}
}
