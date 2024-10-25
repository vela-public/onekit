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
