package jsonkit

import "github.com/vela-public/onekit/cast"

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
