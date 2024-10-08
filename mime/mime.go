package mime

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var (
	mutex sync.RWMutex
)

var (
	DefaultEncode = BinaryEncode
	DefaultDecode = BinaryDecode

	mimeEncode = make(map[string]EncodeFunc)
	mimeDecode = make(map[string]DecodeFunc)

	notFoundEncode = errors.New("not found mime encode")
)

type EncodeFunc func(interface{}) ([]byte, error)
type DecodeFunc func([]byte) (interface{}, error)

func Encode(v interface{}) ([]byte, string, error) {
	name := Name(v)
	switch vt := v.(type) {
	case Encoder:
		data, err := vt.MimeEncode()
		return data, name, err

	default:
		fn := mimeEncode[name]
		if fn == nil {
			fn = DefaultEncode
		}

		data, err := fn(v)
		if err == nil {
			return data, name, nil
		}
		return nil, name, err
	}
}

func Check[T any](data []byte) (T, error) {
	var t T
	var ok bool

	name := Name(t)
	fn := mimeDecode[name]
	if fn == nil {
		fn = DefaultDecode
	}

	v, err := fn(data)
	if err != nil {
		return t, err
	}

	t, ok = v.(T)
	if ok {
		return t, nil
	}
	return t, fmt.Errorf("type mismatch")
}

func Decode(name string, data []byte) (interface{}, error) {
	fn := mimeDecode[name]
	if fn == nil {
		fn = DefaultDecode
	}
	return fn(data)
}

func Name(v interface{}) string {
	if v == nil {
		return "nil"
	}

	vt := reflect.TypeOf(v)

LOOP:
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
		goto LOOP
	}
	return vt.String()
}

func Register(v interface{}, encode EncodeFunc, decode DecodeFunc) {
	mutex.Lock()
	defer mutex.Unlock()
	name := Name(v)
	if _, ok := mimeDecode[name]; ok {
		panic("duplicate mime decode name " + name)
		return
	}

	if _, ok := mimeEncode[name]; ok {
		panic("duplicate mime encode name " + name)
		return
	}

	mimeDecode[name] = decode
	mimeEncode[name] = encode
}

func init() {
	Register(nil, conventionalEncodeFunc, NullDecode)
	Register("", conventionalEncodeFunc, StringDecode)
	Register([]byte{}, conventionalEncodeFunc, BytesDecode)
	Register(true, conventionalEncodeFunc, BoolDecode)
	Register(float64(0), conventionalEncodeFunc, Float64Decode)
	Register(float32(0), conventionalEncodeFunc, Float32Decode)
	Register(int(0), conventionalEncodeFunc, Int32Decode)
	Register(int8(0), conventionalEncodeFunc, Int8Decode)
	Register(int16(0), conventionalEncodeFunc, Int16Decode)
	Register(int32(0), conventionalEncodeFunc, Int32Decode)
	Register(int64(0), conventionalEncodeFunc, Int64Decode)
	Register(uint(0), conventionalEncodeFunc, Uint32Decode)
	Register(uint8(0), conventionalEncodeFunc, Uint8Decode)
	Register(uint16(0), conventionalEncodeFunc, Uint16Decode)
	Register(uint32(0), conventionalEncodeFunc, Uint32Decode)
	Register(uint64(0), conventionalEncodeFunc, Uint64Decode)
	Register(time.Now(), conventionalEncodeFunc, TimeDecode)
}
