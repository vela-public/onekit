package mime

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"strconv"
)

func Int[T int8 | int16 | int32 | int | int64](data []byte, bitSize int) (T, error) {
	n, err := strconv.ParseInt(cast.B2S(data), 10, bitSize)
	return T(n), err
}
func UInt[T uint8 | uint16 | uint32 | uint | uint64](data []byte, bitSize int) (T, error) {
	n, err := strconv.ParseUint(cast.B2S(data), 10, bitSize)
	return T(n), err
}

func FormatInt[T int8 | int16 | int32 | int | int64](a any) ([]byte, error) {
	if v, ok := a.(T); ok {
		return cast.S2B(strconv.FormatInt(int64(v), 10)), nil
	}
	var t T
	return nil, fmt.Errorf("unable encode must:%T got:%T", t, a)
}

func FormatUInt[T uint8 | uint16 | uint32 | uint | uint64](a any) ([]byte, error) {
	if v, ok := a.(T); ok {
		return cast.S2B(strconv.FormatUint(uint64(v), 10)), nil
	}
	var t T
	return nil, fmt.Errorf("unable encode must:%T got:%T", t, a)
}
