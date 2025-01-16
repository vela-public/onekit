package strkit

import (
	"encoding/base64"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/todo"
	"net/url"
	"sort"
)

const (
	StrNum  = string("S")
	String  = string("A")
	Number  = string("N")
	Unicode = string("E")
	Unknown = string("?")
)

type Keywords []string

func (kw Keywords) Len() int {
	return len(kw)
}

func (kw Keywords) Less(i, j int) bool {
	return len(kw[i]) > len(kw[j])
}

func (kw Keywords) Swap(i, j int) {
	kw[i], kw[j] = kw[j], kw[i]
}

func (kw Keywords) MinLen() int {
	if len(kw) > 0 {
		return len(kw[0])
	}
	return -1
}

func (kw Keywords) MaxLen() int {
	if sz := len(kw); sz > 0 {
		return len(kw[sz-1])
	}
	return -1
}

func (kw Keywords) prefix(idx int, total int, text string) (string, bool) {
	n := kw.Len()
	if n == 0 {
		return "", false
	}

	if idx+kw.MinLen() >= total {
		return "", false
	}

	data := cast.S2B(text[idx:])

	comp := func(b []byte) bool {
		size := len(b)
		if idx+size >= total {
			return false
		}

		for i := 0; i < size; i++ {
			v1 := b[i]
			v2 := data[i]

			v1 = todo.IF(v1 >= 'A' && v1 <= 'Z', v1+'a'-'A', v1)
			v2 = todo.IF(v2 >= 'A' && v2 <= 'Z', v2+'a'-'A', v2)
			if v1 != v2 {
				return false
			}
		}
		return true
	}

	for i := n - 1; i >= 0; i-- {
		elem := cast.S2B(kw[i])
		if len(elem) == 0 {
			continue
		}

		if comp(elem) {
			return cast.B2S(elem), true
		}
	}

	return "", false

}

type MaskTag struct {
	Tag  string
	From int
	To   int
}

type SafetyFSM struct {
	Size       int
	Bad        []string
	Other      int
	Dot        int
	Slash      int        // /api/v1/usr/info => 4
	Unknown    int        // not ascii 未知编码 可能是中文
	Alphabetic int        //alphabetic 个数
	Numeric    int        //numeric 个数
	Ext        string     //ext
	Mask       []*MaskTag //掩码字符

	state struct {
		Ext     int
		Current string
		Last    string
		Next    int
	}

	MaskText struct {
		detail []byte
		simple []byte
		norm   []byte
	}
}

type Safety struct {
	short   bool
	sqli    bool
	xss     bool
	codec   []func(string) string
	private struct {
		Bad      *Matcher
		Keywords Keywords
		Drop     []rune //
		Hold     []rune //保留字符
	}
}

func In[T comparable](ch T, data []T) bool {
	if sz := len(data); sz == 0 {
		return false
	} else {
		for i := 0; i < sz; i++ {
			if data[i] == ch {
				return true
			}
		}
	}
	return false
}

func (s *Safety) Short() {
	s.short = true
}

func (s *Safety) Build() {
	if len(s.private.Keywords) != 0 {
		sort.Sort(s.private.Keywords)
	}

	if s.private.Bad != nil {
		s.private.Bad.Build()
	}
}

func (s *Safety) Bad(v ...string) {
	if s.private.Bad == nil {
		s.private.Bad = NewAcMatcher(false)
	}
	if sz := len(v); sz > 0 {
		for i := 0; i < sz; i++ {
			s.private.Bad.Insert(v[i])
		}
	}
}

func (s *Safety) have(idx int, total int, data string) (sub string, ok bool) {
	return s.private.Keywords.prefix(idx, total, data)
}

func (s *Safety) Drop(ch rune) {
	if sz := len(s.private.Drop); sz == 0 {
		s.private.Drop = []rune{ch}
	} else {
		for i := 0; i < sz; i++ {
			if s.private.Drop[i] == ch {
				return
			}
		}
		s.private.Drop = append(s.private.Drop, ch)
	}
}

func (s *Safety) Drops(ch string) {
	for _, v := range ch {
		s.Drop(v)
	}
}

func (s *Safety) SQLi() {
	s.sqli = true
}
func (s *Safety) Xss() {
	s.xss = true
}

func (s *Safety) UnescapeUri() {
	s.codec = append(s.codec, func(v string) string {
		text, err := url.QueryUnescape(v)
		if err != nil {
			return v
		}
		return text
	})
}

func (s *Safety) Base64() {
	s.codec = append(s.codec, func(v string) string {
		text, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return v
		}
		return cast.B2S(text)
	})
}

func (s *Safety) Hold(ch rune) {
	if sz := len(s.private.Hold); sz == 0 {
		s.private.Hold = []rune{ch}
	} else {
		for i := 0; i < sz; i++ {
			if s.private.Hold[i] == ch {
				return
			}
		}
		s.private.Hold = append(s.private.Hold, ch)
	}
}

func (s *Safety) Holds(ch string) {
	for _, v := range ch {
		s.Hold(v)
	}
}

func (s *Safety) Keyword(v ...string) {
	s.private.Keywords = append(s.private.Keywords, v...)
}

func (s *Safety) Decode(v string) string {
	text := v
	if sz := len(s.codec); sz > 0 {
		for i := 0; i < sz; i++ {
			text = s.codec[i](text)
		}
	}
	return text
}

func (s *Safety) Do(text string) *SafetyFSM {
	sz := len(text)
	fsm := &SafetyFSM{
		Size: sz,
	}
	fsm.state.Ext = -1
	if sz == 0 {
		return fsm
	}

	text = s.Decode(text)

	var bad *MatcherFSM
	if s.private.Bad != nil {
		bad = s.private.Bad.Stream()
	}

	for i, ch := range text {
		if bad != nil {
			if term := bad.Input(ch); term != nil {
				fsm.Bad = append(fsm.Bad, text[term.From:term.To])
			}
		}

		if i < fsm.state.Next {
			continue
		}

		if kw, ok := s.have(i, sz, text); ok {
			fsm.state.Next = i + len(kw)
			fsm.Mask = append(fsm.Mask, &MaskTag{
				Tag:  kw,
				From: i,
				To:   i,
			})
			fsm.state.Current = "K"
			fsm.state.Last = "K"
			continue
		}

		if In(ch, s.private.Drop) {
			goto CONTINUE
		}

		if ch == '/' {
			fsm.Slash++
		}

		if ch == '.' {
			fsm.state.Ext = i
			fsm.Dot++
		}

		if In(ch, s.private.Hold) {
			fsm.state.Current = string(ch)
			goto NEXT
		}

		if ch > 127 {
			fsm.Unknown++
			fsm.state.Current = Unicode
			goto NEXT
		}

		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			fsm.Alphabetic++
			if s.short && (fsm.state.Last == StrNum || fsm.state.Last == Number) {
				fsm.state.Current = StrNum
				fsm.Last().Tag = StrNum
			} else {
				fsm.state.Current = String
			}

			goto NEXT
		}

		if ch >= '0' && ch <= '9' {
			fsm.Numeric++
			if s.short && (fsm.state.Last == StrNum || fsm.state.Last == String) {
				fsm.state.Current = StrNum
				fsm.state.Last = StrNum
				fsm.Last().Tag = StrNum
			} else {
				fsm.state.Current = Number
			}
			goto NEXT
		}

		fsm.Other++
		fsm.state.Current = Unknown

	NEXT:
		if fsm.state.Last == fsm.state.Current {
			m := fsm.Last()
			m.To++
			goto CONTINUE
		}

		fsm.Mask = append(fsm.Mask, &MaskTag{
			Tag:  fsm.state.Current,
			From: i,
			To:   i + 1,
		})

		fsm.state.Last = fsm.state.Current

	CONTINUE:
		i++
	}

	if fsm.state.Ext > 0 {
		fsm.Ext = text[fsm.state.Ext:]
	}

	return fsm
}

func NewSafety() *Safety {
	s := &Safety{}
	return s
}
