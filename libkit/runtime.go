package libkit

import "runtime"

func StackTrace[T string | []byte](size int, all bool) T {
	if size < 4096 {
		size = 4096
	}

	buf := make([]byte, size)
	n := runtime.Stack(buf, all)
	return T(buf[:n])
}
