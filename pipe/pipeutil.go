package pipe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"strconv"
)

func MarshalText(item any) ([]byte, error) {
	switch v := item.(type) {
	case string:
		return cast.S2B(v), nil
	case []byte:
		return v, nil
	case Bytes:
		return v.Bytes(), nil
	case int8:
		text := strconv.FormatInt(int64(v), 10)
		return cast.S2B(text), nil
	case int16:
		text := strconv.FormatInt(int64(v), 10)
		return cast.S2B(text), nil
	case int32:
		text := strconv.FormatInt(int64(v), 10)
		return cast.S2B(text), nil
	case int:
		text := strconv.FormatInt(int64(v), 10)
		return cast.S2B(text), nil
	case int64:
		text := strconv.FormatInt(v, 10)
		return cast.S2B(text), nil
	case uint8:
		text := strconv.FormatUint(uint64(v), 10)
		return cast.S2B(text), nil
	case uint16:
		text := strconv.FormatUint(uint64(v), 10)
		return cast.S2B(text), nil
	case uint32:
		text := strconv.FormatUint(uint64(v), 10)
		return cast.S2B(text), nil
	case uint:
		text := strconv.FormatUint(uint64(v), 10)
		return cast.S2B(text), nil
	case uint64:
		text := strconv.FormatUint(uint64(v), 10)
		return cast.S2B(text), nil

	case float32:
		text := strconv.FormatFloat(float64(v), 'f', -1, 64)
		return cast.S2B(text), nil
	case float64:
		text := strconv.FormatFloat(v, 'f', -1, 64)
		return cast.S2B(text), nil
	case bool:
		text := strconv.FormatBool(v)
		return cast.S2B(text), nil
	case fmt.Stringer:
		return cast.S2B(v.String()), nil
	case bytes.Buffer:
		return v.Bytes(), nil
	case *bytes.Buffer:
		return v.Bytes(), nil

	default:
		text, err := json.Marshal(item)
		return text, err
	}
}
