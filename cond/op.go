package cond

const (
	Eq op = iota + 10
	Re
	Cn
	In
	Lt
	Le
	Ge
	Gt
	Unary
	Call
	Oop
	Pass
	Regex
	Cidr
	Fn
)

var (
	opTab = []string{"equal", "grep", "contain", "include", "less", "less or equal", "greater or equal", "greater", "unary", "call", "oop", "pass", "regex"}
)

var (
	opText = map[string]bool{
		"==":  true,
		"eq":  true,
		"re":  true,
		"cn":  true,
		"in":  true,
		"lt":  true,
		"gt":  true,
		"le":  true,
		"<=":  true,
		"ge":  true,
		">=":  true,
		"->":  true,
		"ieq": true,
		"icn": true,
		"iin": true,
		"ire": true,
		"~":   true,
		"=":   true,
		">":   true,
		"<":   true,
	}
)

type op uint8

func (o op) String() string {
	return opTab[(int(o) - 10)]
}
