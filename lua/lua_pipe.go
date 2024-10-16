package lua

func pipeL(L *LState) int {
	if L.Pipe == nil {
		L.Push(LNil)
	} else {
		L.Push(L.Pipe)
	}
	return 1
}
