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

type op uint8

func (o op) String() string {
	return opTab[(int(o) - 10)]
}
