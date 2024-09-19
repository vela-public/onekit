package jsonkit

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/bytebufferpool"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
	"strconv"
	"time"
)

type ByteBuffer = bytebufferpool.ByteBuffer

type JsonBuffer struct {
	buffer *ByteBuffer
}

func NewJson() *JsonBuffer {
	buff := bytebufferpool.Get()
	return &JsonBuffer{buffer: buff}
}

func (j *JsonBuffer) Char(ch byte) {
	j.buffer.WriteByte(ch)
}

func (j *JsonBuffer) WriteByte(ch byte) {
	switch ch {
	case '\\':
		j.buffer.WriteString("\\\\")
	case '\r':
		j.buffer.WriteString("\\r")

	case '\n':
		j.buffer.WriteString("\\n")

	case '\t':
		j.buffer.WriteString("\\t")
	case '"':
		j.buffer.WriteString("\\\"")

	default:
		j.buffer.WriteByte(ch)
	}
}

func (j *JsonBuffer) WriteString(val string) {
	n := len(val)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		j.WriteByte(val[i])
	}
}

func (j *JsonBuffer) Write(val []byte) {
	n := len(val)
	if n == 0 {
		return
	}

	for i := 0; i < n; i++ {
		j.WriteByte(val[i])
	}
}

func (j *JsonBuffer) Key(key string) {
	j.Char('"')
	j.WriteString(key)
	j.Char('"')
	j.WriteByte(':')
}

func (j *JsonBuffer) Val(v string) {
	j.Char('"')
	j.WriteString(v)
	j.Char('"')
}

func (j *JsonBuffer) Insert(v []byte) {
	j.Char('"')
	j.Write(v)
	j.Char('"')
}

func (j *JsonBuffer) Int(n int) {
	j.WriteString(strconv.Itoa(n))
}

func (j *JsonBuffer) Bool(v bool) {
	if v {
		j.Write(True)
	} else {
		j.Write(False)
	}
}

func (j *JsonBuffer) Long(n int64) {
	j.WriteString(cast.ToString(n))
}

func (j *JsonBuffer) ULong(n uint64) {
	j.WriteString(cast.ToString(n))
}

func (j *JsonBuffer) KT(key string, t time.Time) {
	j.Key(key)
	j.Val(t.String())
	j.WriteByte(',')
}

//func (enc *JsonEncoder) KV(key , val string) {
//	enc.Key(key)
//	enc.Val(val)
//	enc.WriteByte(',')
//}

func (j *JsonBuffer) KI(key string, n int) {
	j.Key(key)
	j.Int(n)
	j.WriteByte(',')
}

func (j *JsonBuffer) ToStr(key string, v string) {
	j.kv2(key, v)
}

func (j *JsonBuffer) ToBytes(key string, v []byte) {
	j.kv2(key, cast.B2S(v))
}

func (j *JsonBuffer) KF64(key string, v float64) {
	j.Key(key)
	j.WriteString(cast.ToString(v))
	j.WriteByte(',')
}

func (j *JsonBuffer) KL(key string, n int64) {
	j.Key(key)
	j.Long(n)
	j.WriteByte(',')
}

func (j *JsonBuffer) KUL(key string, n uint64) {
	j.Key(key)
	j.ULong(n)
	j.WriteByte(',')
}

func (j *JsonBuffer) Join(key string, v []string) {
	j.Key(key)

	j.Arr("")
	for _, item := range v {
		j.Val(item)
		j.WriteByte(',')
	}

	j.End("]")
	j.WriteByte(',')
}

func (j *JsonBuffer) NoKeyJoin(v []string) {
	j.Arr("")
	for _, item := range v {
		j.Val(item)
		j.WriteByte(',')
	}

	j.End("]")
	j.WriteByte(',')
}

func (j *JsonBuffer) NoKeyJoin2(v []interface{}) {
	j.Arr("")
	for _, item := range v {
		j.WriteString(cast.ToString(item))
		j.WriteByte(',')
	}

	j.End("]")
	j.WriteByte(',')
}

func (j *JsonBuffer) Join2(key string, v []interface{}) {
	j.Key(key)

	j.Arr("")
	for _, item := range v {
		j.WriteString(cast.ToString(item))
		j.WriteByte(',')
	}

	j.End("]")
	j.WriteByte(',')
}

func (j *JsonBuffer) kv1(key, v string) {
	j.Key(key)
	j.WriteString(v)
	j.WriteByte(',')
}

func (j *JsonBuffer) kv2(key, v string) {
	j.Key(key)
	j.Val(v)
	j.WriteByte(',')
}

func (j *JsonBuffer) V1(v string) {
	j.WriteString(v)
	j.WriteByte(',')
}

func (j *JsonBuffer) V2(v string) {
	j.Val(v)
	j.WriteByte(',')
}

func (j *JsonBuffer) V(v interface{}) {
	switch val := v.(type) {
	case nil:
		j.V2("")

	case bool:
		j.V1(strconv.FormatBool(val))
	case float64:
		j.V1(strconv.FormatFloat(val, 'f', -1, 64))
	case float32:
		j.V1(strconv.FormatFloat(float64(val), 'f', -1, 64))
	case int:
		j.V1(strconv.Itoa(val))
	case int64:
		j.V1(strconv.FormatInt(val, 10))
	case int32:
		j.V1(strconv.Itoa(int(val)))

	case int16:
		j.V1(strconv.FormatInt(int64(val), 10))
	case int8:
		j.V1(strconv.FormatInt(int64(val), 10))
	case uint:
		j.V1(strconv.FormatUint(uint64(val), 10))
	case uint64:
		j.V1(strconv.FormatUint(val, 10))
	case uint32:
		j.V1(strconv.FormatUint(uint64(val), 10))
	case uint16:
		j.V1(strconv.FormatUint(uint64(val), 10))
	case uint8:
		j.V1(strconv.FormatUint(uint64(val), 10))

	case string:
		j.V2(val)

	case lua.LString:
		j.V2(string(val))
	case lua.LBool:
		j.V1(strconv.FormatBool(bool(val)))
	case lua.LNilType:
		j.V2("nil")

	case lua.LNumber:
		j.V1(strconv.FormatFloat(float64(val), 'f', -1, 64))
	case lua.LInt:
		j.V1(strconv.Itoa(int(val)))

	case []string:
		j.NoKeyJoin(val)
	case []byte:
		j.V2(cast.B2S(val))

	case []interface{}:
		j.NoKeyJoin2(val)

	case time.Time:
		if y := val.Year(); y < 0 || y >= 10000 {
			// RFC 3339 is clear that years are 4 digits exactly.
			// See golang.org/issue/4556#c15 for more discussion.

			return
			//return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
		}
		j.V2(val.Format(time.RFC3339Nano))
	case error:
		j.V2(val.Error())

	default:
		chunk, err := json.Marshal(val)
		if err != nil {
			return
		}
		j.buffer.Write(chunk)
		j.buffer.WriteByte(',')
	}

}

func (j *JsonBuffer) KV(key string, s interface{}) {
	switch val := s.(type) {
	case nil:
		j.kv2(key, "")

	case bool:
		j.kv1(key, strconv.FormatBool(val))
	case float64:
		j.kv1(key, strconv.FormatFloat(val, 'f', -1, 64))
	case float32:
		j.kv1(key, strconv.FormatFloat(float64(val), 'f', -1, 64))
	case int:
		j.kv1(key, strconv.Itoa(val))
	case int64:
		j.kv1(key, strconv.FormatInt(val, 10))
	case int32:
		j.kv1(key, strconv.Itoa(int(val)))

	case int16:
		j.kv1(key, strconv.FormatInt(int64(val), 10))
	case int8:
		j.kv1(key, strconv.FormatInt(int64(val), 10))
	case uint:
		j.kv1(key, strconv.FormatUint(uint64(val), 10))
	case uint64:
		j.kv1(key, strconv.FormatUint(val, 10))
	case uint32:
		j.kv1(key, strconv.FormatUint(uint64(val), 10))
	case uint16:
		j.kv1(key, strconv.FormatUint(uint64(val), 10))
	case uint8:
		j.kv1(key, strconv.FormatUint(uint64(val), 10))

	case string:
		j.kv2(key, val)

	case lua.LString:
		j.kv2(key, string(val))
	case lua.LBool:
		j.kv1(key, strconv.FormatBool(bool(val)))
	case lua.LNilType:
		j.kv2(key, "nil")

	case lua.LNumber:
		j.kv1(key, strconv.FormatFloat(float64(val), 'f', -1, 64))
	case lua.LInt:
		j.kv1(key, strconv.Itoa(int(val)))

	case []string:
		Join[string](j, key, val, true)
	case []byte:
		j.kv2(key, cast.B2S(val))
	case []bool:
		Join[bool](j, key, val, false)

	case []int:
		Join[int](j, key, val, false)
	case []float64:
		Join[float64](j, key, val, false)
	case []interface{}:
		chunk, err := json.Marshal(val)
		if err != nil {
			j.Raw(key, EmptyA)
			return
		}
		j.Raw(key, chunk)

	case time.Time:
		if y := val.Year(); y < 0 || y >= 10000 {
			// RFC 3339 is clear that years are 4 digits exactly.
			// See golang.org/issue/4556#c15 for more discussion.

			return
			//return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
		}
		j.kv2(key, val.Format(time.RFC3339Nano))
	case error:
		j.kv2(key, val.Error())

	case fmt.Stringer:
		j.kv2(key, val.String())

	default:
		chunk, err := json.Marshal(val)
		if err != nil {
			j.kv2(key, err.Error())
			return
		}
		j.Raw(key, chunk)
	}

}

var False = []byte("false")
var True = []byte("true")

func (j *JsonBuffer) KB(key string, b bool) {
	j.Key(key)

	if b {
		j.Write(True)
	} else {
		j.Write(False)
	}

	j.WriteByte(',')
}

func (j *JsonBuffer) False(key string) {
	j.Key(key)
	j.Write(False)
	j.WriteByte(',')
}

func (j *JsonBuffer) True(key string) {
	j.Key(key)
	j.Write(True)
	j.WriteByte(',')
}

func (j *JsonBuffer) Tab(name string) {
	if len(name) != 0 {
		j.Val(name)
		j.WriteByte(':')
	}

	j.WriteByte('{')
}

func (j *JsonBuffer) Arr(name string) {
	if len(name) != 0 {
		j.Val(name)
		j.WriteByte(':')
	}
	j.WriteByte('[')
}

func (j *JsonBuffer) Append(val []byte) {
	n := len(val)
	if n == 0 {
		return
	}
	j.buffer.Write(val)
	j.buffer.WriteByte(',')
}

func (j *JsonBuffer) Raw(key string, val []byte) {
	n := len(val)
	if n == 0 {
		return
	}

	j.Key(key)
	j.buffer.Write(val)
	j.WriteByte(',')
}

func (j *JsonBuffer) Copy(val []byte) {
	if len(val) == 0 {
		return
	}
	j.buffer.Write(val)
}

func (j *JsonBuffer) Marshal(key string, v interface{}) error {
	if v == nil {
		return fmt.Errorf("nil value")
	}
	chunk, err := json.Marshal(v)
	if err != nil {
		return err
	}
	j.Raw(key, chunk)
	return nil

}

func (j *JsonBuffer) End(val string) {
	n := j.buffer.Len()

	if n <= 0 {
		return
	}

	if j.buffer.B[n-1] == ',' {
		j.buffer.B = j.buffer.B[:n-1]
	}

	j.WriteString(val)
}

func (j *JsonBuffer) Bytes() []byte {
	return j.buffer.Bytes()
}

func (j *JsonBuffer) Buffer() *ByteBuffer {
	return j.buffer
}
