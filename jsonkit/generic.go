package jsonkit

import (
	"github.com/vela-public/onekit/cast"
	"strconv"
	"strings"
)

var EmptyA = []byte("[]")

func Join[T any](enc *JsonBuffer, key string, data []T, quote bool) {
	n := len(data)
	if n == 0 {
		enc.Raw(key, EmptyA)
		return
	}

	enc.Key(key)
	enc.Arr("")

	for i := 0; i < n; i++ {
		item := data[i]
		if quote {
			enc.Val(cast.ToString(item))
			enc.WriteByte(',')
			continue
		}

		enc.WriteString(cast.ToString(item))
		enc.WriteByte(',')
	}
	enc.End("]")
	enc.WriteByte(',')
}

func Unquote(value string) string {
	size := len(value)
	if size == 0 {
		return ""
	}

	if size >= 2 && value[0] == '"' && value[size-1] == '"' {
		if strings.Index(value, "\\") == -1 {
			return value[1 : size-1]
		}

		text, err := strconv.Unquote(value)
		if err != nil {
			return ""
		}
		return text
	}

	return value
}

func Quote(value string) string {
	sz := len(value)
	if sz == 0 {
		return "\"\""
	}

	if sz >= 2 && value[0] == '"' && value[sz-1] == '"' {
		return value
	}

	return strconv.Quote(value)
}
