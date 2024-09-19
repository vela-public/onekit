package errkit

import (
	"github.com/vela-public/onekit/cast"
	"runtime"
)

func StackTrace(size int) string {
	var buf []byte
	if size == 0 {
		buf = make([]byte, 8192)
	} else {
		buf = make([]byte, size)
	}

	n := runtime.Stack(buf[:], false)
	return cast.B2S(buf[:n])
}
