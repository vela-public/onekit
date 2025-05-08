package lua

const (
	StructLibName = "struct"
)

func NewStructL(L *LState) int {
	return 0
}

func structL(L *LState, key string) LValue {
	return LNil
}

func OpenStructLib(L *LState) int {
	mod := NewExport("lua.struct.export", WithFunc(NewStructL), WithIndex(structL))
	L.SetGlobal("struct", mod)
	L.Push(mod)
	return 1

}

/*

   local struct = vela.def[[
		name string
        age int
        sex bool
        height float
        weight double
	]]

	local data = struct{
		"name" = "string",
		"age" = "int",
		"sex" = "bool",
		"height" = "float",
		"weight" = "double"
	}




*/
