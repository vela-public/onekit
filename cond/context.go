package cond

type CndContext struct {
	Value   any
	Lookup  Lookup
	handler struct {
		Error func(error)
		Debug func(any, *Section, bool)
	}
}
