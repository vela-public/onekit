package abi

import (
	"bytes"
	"github.com/vela-public/onekit/cast"
	"unsafe"
)

func BytesToCleanString(b []byte) string {
	i := bytes.LastIndexFunc(b, func(r rune) bool {
		return r == 0x00
	})

	if i != -1 {
		return cast.B2S(b[:i])
	}
	return cast.B2S(b)
}

func Cast[T any](s *StructInstance) (t T, ok bool) {
	if s == nil {
		return t, false
	}

	raw := s.buffer
	if len(raw) < int(unsafe.Sizeof(t)) {
		return t, false
	}
	return *(*T)(unsafe.Pointer(&raw[0])), true
}
