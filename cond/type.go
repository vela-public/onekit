package cond

type Method func(string, string) bool

type CompareEx interface {
	Compare(string, string, Method) bool //key string , val string , equal
}

type Lookup func(string) string

type Retrieval struct {
	Value  interface{}
	Lookup Lookup
}
