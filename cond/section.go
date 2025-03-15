package cond

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/grep"
	"github.com/vela-public/onekit/lua"
	"net"
	"regexp"
	"strings"
)

type Section struct {
	err       error
	not       bool
	iCase     bool
	method    op
	raw       string
	keys      []string
	data      []string
	regex     []*regexp.Regexp
	subnet    []*net.IPNet
	partition int
	invoke    func(any, ...OptionFunc) bool
}

func (s *Section) Partition(part int) {
	s.partition = part
}

func (s *Section) WithNot() {
	s.not = true
}

func (s *Section) Regex(v ...string) {
	for _, item := range v {
		r := regexp.MustCompile(item)
		if r == nil {
			continue
		}
		s.data = append(s.data, item)
		s.regex = append(s.regex, r)
	}
}

func (s *Section) Method(v op) {
	s.method = v
}

func (s *Section) Keys(v ...string) {
	s.keys = append(s.keys, v...)
}

func (s *Section) Value(v ...string) {
	s.data = append(s.data, v...)
}

func (s *Section) invalid(format string, v ...interface{}) {
	s.err = fmt.Errorf(format, v...)
}

func (s *Section) trim(offset *int, n int) {
	for i := *offset; i < n; i++ {
		if ch := s.raw[i]; ch != ' ' {
			*offset = i
			return
		}
	}
}

func (s *Section) Ok() bool {
	return s.err == nil
}

func (s *Section) withA(offset *int, n int) { //"name , key , text =  vela,123,456"
	s.trim(offset, n)
	sep := *offset

	for i := *offset; i < n; i++ {
		ch := s.raw[i]
		switch ch {
		case ',':
			s.keys = append(s.keys, s.raw[sep:i])
			sep = i

		case ' ':
			if sep != i {
				s.keys = append(s.keys, s.raw[sep:i])
			}
			*offset = i
			return
		}
	}

}

func (s *Section) withB(offset *int, n int) {
	if !s.Ok() {
		return
	}

	s.trim(offset, n)
	sep := *offset

	if sep+3 > n {
		s.invalid("not found method")
		return
	}

	if s.raw[sep] == '!' {
		s.not = true
		sep++
	}

	switch s.raw[sep] {
	case '~':
		s.method = Regex
		*offset = sep + 1
		return
	case '=':
		s.method = Eq
		if sep+1 < len(s.raw) && s.raw[sep+1] == '=' {
			*offset = sep + 2
			return
		}
		*offset = sep + 1
		return

	case '>':
		if sep+1 < len(s.raw) && s.raw[sep+1] == '=' {
			*offset = sep + 2
			s.method = Ge
			return
		}
		s.method = Gt
		*offset = sep + 1
		return

	case '<':
		if sep+1 < len(s.raw) && s.raw[sep+1] == '=' {
			*offset = sep + 2
			s.method = Le
			return
		}
		s.method = Lt
		*offset = sep + 1
		return
	}

	em := s.raw[sep : sep+2]
	switch em {
	case "==":
		s.method = Eq
		*offset = sep + 2
		return
	case "eq":
		s.method = Eq
		*offset = sep + 2
		return
	case "re":
		s.method = Re
		*offset = sep + 2
		return
	case "cn":
		s.method = Cn
		*offset = sep + 2
		return
	case "in":
		s.method = In
		*offset = sep + 2
		return
	case "lt":
		s.method = Lt
		*offset = sep + 2
		return
	case "gt":
		s.method = Gt
		*offset = sep + 2
		return
	case "le", "<=":
		s.method = Le
		*offset = sep + 2
		return
	case "ge", ">=":
		s.method = Ge
		*offset = sep + 2
		return
	case "->":
		s.method = Call
		*offset = sep + 2
		return

	}

	em = s.raw[sep : sep+3]
	switch em {
	case "ieq":
		s.method = Eq
		s.iCase = true
		*offset = sep + 3
		return
	case "icn":
		s.method = Cn
		s.iCase = true
		*offset = sep + 3
		return
	case "iin":
		s.method = In
		s.iCase = true
		*offset = sep + 3
		return
	case "ire":
		s.method = Re
		s.iCase = true
		*offset = sep + 3
		return
	}

}

func (s *Section) withC(offset *int, n int) {
	if !s.Ok() {
		return
	}

	s.trim(offset, n)
	sep := *offset
	var item string
	for i := *offset; i < n; i++ {
		ch := s.raw[i]
		if ch != ',' {
			continue
		}

		if s.raw[i-1] == '\\' {
			continue
		}

		if s.raw[sep] == ',' {
			item = s.raw[sep+1 : i]
		} else {
			item = s.raw[sep:i]
		}

		s.Value(item)
		sep = i
	}

	//single value
	if sep == *offset {
		s.Value(s.raw[sep:])
		return
	}

	//last value
	if sep != n-1 {
		s.Value(s.raw[sep+1:])
	}
}

func (s *Section) re2() {
	if s.method != Regex {
		return
	}

	if len(s.data) == 0 {
		s.err = fmt.Errorf("not found data")
		return
	}

	for _, item := range s.data {
		r, err := regexp.Compile(item)
		if r == nil {
			s.err = err
			return
		}

		s.data = append(s.data, item)
		s.regex = append(s.regex, r)
	}
}

func (s *Section) compare(a string, b string) bool {

	result := false
	switch s.method {
	case Eq:
		if a == "" && b == "nil" {
			result = true
			goto done
		}
		if s.iCase {
			result = strings.EqualFold(a, b)
		} else {
			result = a == b
		}

		goto done

	case Re:
		result = grep.New(b)(a)
		goto done
	case Cn:
		if s.iCase {
			result = strings.Contains(strings.ToLower(a), strings.ToLower(b))
		} else {
			result = strings.Contains(a, b)
		}
		goto done
	case In:
		result = a == b
		goto done
	case Lt:
		result = cast.ToFloat64(a) < cast.ToFloat64(b)
		goto done
	case Le:
		result = cast.ToFloat64(a) <= cast.ToFloat64(b)
		goto done
	case Ge:
		result = cast.ToFloat64(a) >= cast.ToFloat64(b)
		goto done
	case Gt:
		result = cast.ToFloat64(a) > cast.ToFloat64(b)
		goto done
	case Unary:
		switch a {
		case "true":
			result = true
		case "false":
			result = false
		case "nil":
			result = false
		case "":
			result = false
		case "0":
			result = false
		default:
			result = true
		}
	default:
		result = false
	}

done:
	return result
}

func (s *Section) newMatch(i int, ov *option) func(string, string) bool {
	if s.method != Regex {
		return func(k string, v string) bool {
			if !s.compare(k, v) {
				return false
			}
			return true
		}
	}

	return func(v string, raw string) bool {
		r := s.regex[i]
		ret := r.FindStringSubmatch(v)
		if len(ret) == 0 {
			return false
		}

		if s.partition >= 1 && s.partition <= len(ret) {
			ov.Pay(i, ret[s.partition-1])
		}

		n := len(ov.partition)
		if n > 0 {
			for ii := 0; ii < n; ii++ {
				part := ov.partition[ii]
				if part >= 1 && part <= len(ret) {
					ov.Pay(i, ret[part-1])
				}
			}
			return true
		}

		if ov.payload != nil {
			for pos, item := range ret {
				ov.Pay(pos, item)
			}
		}

		return true
	}
}

func (s *Section) ContainNet(v string) bool {
	ip := net.ParseIP(v)
	if ip == nil {
		return false
	}

	if len(s.subnet) == 0 {
		return false
	}

	for _, sub := range s.subnet {
		if sub.Contains(ip) {
			return true
		}
	}

	return false
}

func (s *Section) Match(v string, ov *option) bool {
	switch s.method {
	case Cidr:
		return s.ContainNet(v)
	default:
		n := len(s.data)
		for i := 0; i < n; i++ {
			item := s.data[i]
			fn := s.newMatch(i, ov)
			if fn(v, item) {
				return true
			}
		}
	}

	return false
}

func (s *Section) Compare(ov *option, v string) bool {
	n := len(s.data)
	for i := 0; i < n; i++ {
		fn := s.newMatch(i, ov)
		if ov.compare(v, s.data[i], fn) {
			return true
		}
	}
	return false
}

func (s *Section) Unary(ov *option) (bool, error) { // unary: section:{data:[]string(data1 , data2))
	if ov.value == nil {
		return false, nil
	}

	n := len(s.data)
	if n == 0 {
		return false, nil
	}

	str, err := cast.ToStringE(ov.value)
	if err != nil {
		return false, err
	}

	for i := 0; i < n; i++ {
		if strings.Contains(str, s.data[i]) {
			return true, nil
		}
	}

	return false, nil
}

func (s *Section) pure(ov *option) (bool, error) {
	n := len(s.keys)
	for i := 0; i < n; i++ {
		if ov.compare != nil {
			if s.Compare(ov, s.keys[i]) {
				return s.not != true, s.err
			}
			continue
		}

		if !s.Match(ov.field(s.keys[i]), ov) {
			continue
		}

		return s.not != true, s.err
	}
	return s.not != false, s.err
}

func (s *Section) Call(ov *option) (bool, error) {
	if ov.field == nil && ov.compare == nil {
		return false, fmt.Errorf("invalid field function")
	}

	if !s.Ok() {
		return false, s.err
	}

	switch {
	case s.method == Pass:
		return true, nil
	case s.method == Fn:
		if s.invoke == nil {
			return false, nil
		}
		return s.invoke(ov.value), nil

	case s.method == Unary && len(s.keys) > 0: //!key , true , false 这类单目运算
		return s.pure(ov)
	case s.method == Unary && len(s.data) > 0: // 单目运算全局匹配
		return s.Unary(ov)
	default:
		return s.pure(ov)
	}

}

func (s *Section) is3(offset int, a, b, c byte) bool {
	if offset-1 < 0 {
		return false
	}

	if offset+1 >= len(s.raw) {
		return false
	}

	if s.raw[offset-1] != a {
		return false
	}
	if s.raw[offset] != b {
		return false
	}
	if s.raw[offset+1] != c {
		return false
	}
	return true
}

// Compile
// aaa eq abc,eee,fff => Section{not:false , keys: []string{aaa} , method: eq , data: []string{abc , eee , ff}}
// aaa !eq abc,eee,fff => Section{not:true, keys: []string{aaa} , method: eq , data: []string{abc , eee , ff}}
func (s *Section) isUnary() bool {
	sz := len(s.raw)
	for i := 0; i < sz; i++ {
		ch := s.raw[i]
		switch ch {
		case '=', '>', '<', '~':
			return false
		case 'e':
			if s.is3(i, ' ', 'e', 'q') {
				return false
			}
			if s.is3(i, '!', 'e', 'q') {
				return false
			}

		case 'i':
			if s.is3(i, ' ', 'i', 'n') {
				return false
			}
			if s.is3(i, '!', 'i', 'n') {
				return false
			}
		case 'c':
			if s.is3(i, ' ', 'c', 'n') {
				return false
			}

			if s.is3(i, '!', 'c', 'n') {
				return false
			}
		case 'l':
			if s.is3(i, ' ', 'l', 't') {
				return false
			}
			if s.is3(i, '!', 'l', 't') {
				return false
			}
			if s.is3(i, ' ', 'l', 'e') {
				return false
			}
			if s.is3(i, '!', 'l', 'e') {
				return false
			}
		case 'g':
			if s.is3(i, ' ', 'g', 't') {
				return false
			}
			if s.is3(i, '!', 'g', 't') {
				return false
			}

		case 'r':
			if s.is3(i, ' ', 'r', 'e') {
				return false
			}

			if s.is3(i, '!', 'r', 'e') {
				return false
			}
		}

	}

	if strings.IndexFunc(s.raw, func(r rune) bool {
		switch r {
		case '=':
			return true
		case '>':
			return true
		case ',':
			return true
		case ' ':
			return true
		default:
			return false
		}
	}) != -1 {
		return false
	}

	if len(s.raw) == 0 {
		return false
	}

	key := s.raw
	if s.raw[0] == '!' {
		s.not = true
		key = s.raw[1:]
	}

	s.keys = append(s.keys, key)
	s.method = Unary
	s.data = append(s.data, "")
	return true
}

func (s *Section) isPassMatch() bool {
	switch s.raw {
	case "", "*":
		s.method = Pass
		return true
	}

	return false
}

func (s *Section) compile() {
	n := len(s.raw)
	if n < 6 {
		s.err = fmt.Errorf("too short")
		return
	}

	offset := 0
	s.withA(&offset, n)
	s.withB(&offset, n)
	s.withC(&offset, n)
	s.re2()
}

func NewSectionText(raw string) (section *Section) {
	section = &Section{
		raw:       strings.ToLower(strings.TrimSpace(raw)),
		method:    Oop,
		partition: -1,
	}

	if section.isPassMatch() {
		return
	}

	if section.isUnary() {
		return
	}

	section.compile()
	return
}

func NewSectionGoFunc(L *lua.LState, invoke func(interface{}, ...OptionFunc) bool) (section *Section) {
	section = &Section{
		raw:       strings.ToLower(strings.TrimSpace(L.String())),
		method:    Fn,
		partition: -1,
	}
	section.invoke = invoke
	return
}

func NewSectionLFunc(L *lua.LState, fn *lua.LFunction) (section *Section) {
	section = &Section{
		raw:       strings.ToLower(fn.String()),
		method:    Fn,
		partition: -1,
	}

	section.invoke = func(v interface{}, optionFunc ...OptionFunc) bool {
		np := lua.P{
			Fn:      fn,
			Protect: true,
			NRet:    1,
		}

		co := L.Coroutine()
		defer func() {
			L.Keepalive(co)
		}()

		err := co.CallByParam(np, lua.ReflectTo(v))
		if err != nil {
			return false
		}
		return lua.IsTrue(co.Get(-1))

	}
	return
}

func NewSectionUnary(data ...string) (section *Section) {
	section = &Section{
		method:    Unary,
		partition: -1,
	}

	section.data = append(section.data, data...)
	return
}
