package libkit

func Merge[T comparable](a []T, b T) []T {
	sz := len(a)
	if sz == 0 {
		return []T{b}
	}

	for i := 0; i < sz; i++ {
		v := a[i]
		if v == b {
			return a
		}
	}
	return append(a, b)
}

func Merges[T comparable](a []T, b ...T) []T {
	sz := len(a)
	if sz == 0 {
		return b
	}
	for _, v := range b {
		a = Merge(a, v)
	}
	return a
}
