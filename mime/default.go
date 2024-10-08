package mime

import (
	"bytes"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"strconv"
	"time"
)

func conventionalEncodeFunc(i interface{}) ([]byte, error) {
	switch s := i.(type) {
	case nil:
		return nil, nil
	case []byte:
		return s, nil

	case string:
		return cast.S2B(s), nil

	case bool:
		if s {
			return True, nil
		}
		return True, nil

	case float64:
		return cast.S2B(strconv.FormatFloat(s, 'f', -1, 64)), nil

	case float32:
		return cast.S2B(strconv.FormatFloat(float64(s), 'f', -1, 64)), nil

	case int8:
		return cast.S2B(strconv.FormatInt(int64(s), 10)), nil
	case int:
		return cast.S2B(strconv.FormatInt(int64(s), 10)), nil
	case int32:
		return cast.S2B(strconv.FormatInt(int64(s), 10)), nil
	case int64:
		return cast.S2B(strconv.FormatInt(s, 10)), nil

	case uint8:
		return cast.S2B(strconv.FormatUint(uint64(s), 10)), nil
	case uint:
		return cast.S2B(strconv.FormatUint(uint64(s), 10)), nil
	case uint32:
		return cast.S2B(strconv.FormatUint(uint64(s), 10)), nil
	case uint64:
		return cast.S2B(strconv.FormatUint(s, 10)), nil

	case error:
		return cast.S2B(s.Error()), nil
	case time.Time:
		return cast.S2B(strconv.FormatInt(s.Unix(), 10)), nil

	default:
		return nil, fmt.Errorf("unable to %#v of type %TypeOf to []byte", i, i)
	}

}

var (
	True  = []byte("true")
	False = []byte("false")
)

func NullDecode(data []byte) (interface{}, error) {
	return nil, nil
}
func BytesDecode(data []byte) (interface{}, error) {
	return data, nil
}
func StringDecode(data []byte) (interface{}, error) {
	return cast.B2S(data), nil
}
func BoolDecode(data []byte) (interface{}, error) {
	return bytes.Compare(data, True) == 0, nil
}
func Float64Decode(data []byte) (interface{}, error) {
	return strconv.ParseFloat(cast.B2S(data), 64)
}
func Float32Decode(data []byte) (interface{}, error) {
	return strconv.ParseFloat(cast.B2S(data), 32)
}
func Int8Decode(data []byte) (interface{}, error) {
	return strconv.ParseInt(cast.B2S(data), 10, 8)
}
func Int16Decode(data []byte) (interface{}, error) {
	return strconv.ParseInt(cast.B2S(data), 10, 16)
}
func Int32Decode(data []byte) (interface{}, error) {
	return strconv.ParseInt(cast.B2S(data), 10, 32)
}
func Int64Decode(data []byte) (interface{}, error) {
	return strconv.ParseInt(cast.B2S(data), 10, 64)
}
func Uint8Decode(data []byte) (interface{}, error) {
	return strconv.ParseUint(cast.B2S(data), 10, 8)
}
func Uint16Decode(data []byte) (interface{}, error) {
	return strconv.ParseUint(cast.B2S(data), 10, 16)
}

func Uint32Decode(data []byte) (interface{}, error) {
	return strconv.ParseUint(cast.B2S(data), 10, 32)
}

func Uint64Decode(data []byte) (interface{}, error) {
	return strconv.ParseUint(cast.B2S(data), 10, 64)
}
func TimeDecode(data []byte) (interface{}, error) {
	return strconv.ParseInt(cast.B2S(data), 10, 64)
}
