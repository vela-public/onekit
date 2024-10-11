package mime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/todo"
	"strconv"
	"time"
)

var (
	Null  = []byte("nil")
	True  = []byte("true")
	False = []byte("false")
)

type Unknown[T any] struct{}

func (u Unknown[T]) MimeDecode(data []byte) (any, error) {
	var t T
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&t)
	return t, err
}

func (u Unknown[T]) MimeEncode(v any) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}

func (u Unknown[T]) TypeFor() interface{} {
	var t T
	return t
}

type Nil struct{}

func (Nil) TypeFor() interface{}                   { return nil }
func (Nil) MimeEncode(interface{}) ([]byte, error) { return Null, nil }
func (Nil) MimeDecode(data []byte) (interface{}, error) {
	if bytes.Compare(data, Null) == 0 {
		return nil, nil
	}
	return nil, fmt.Errorf("unable to decode %s ", string(data))
}

type Text struct{}

func (Text) TypeFor() any                        { return "" }
func (Text) MimeDecode(data []byte) (any, error) { return cast.B2S(data), nil }
func (Text) MimeEncode(a any) ([]byte, error) {
	data, ok := a.(string)
	if ok {
		return cast.S2B(data), nil
	}
	return nil, fmt.Errorf("unable encode must string , got: %T", a)
}

type Bytes struct{}

func (b Bytes) TypeFor() any                        { return []byte{} }
func (b Bytes) MimeDecode(data []byte) (any, error) { return data, nil }
func (b Bytes) MimeEncode(a any) ([]byte, error) {
	data, ok := a.([]byte)
	if ok {
		return data, nil
	}
	return nil, fmt.Errorf("unable encode must []byte , got: %T", a)
}

type Bool struct{}

func (Bool) TypeFor() any { return false }
func (Bool) MimeDecode(data []byte) (any, error) {
	if bytes.Compare(data, True) == 0 {
		return true, nil
	}

	if bytes.Compare(data, False) == 0 {
		return true, nil
	}

	return false, fmt.Errorf("unable to decode %s ", string(data))
}

func (Bool) MimeEncode(a any) ([]byte, error) {
	v, ok := a.(bool)
	if ok {
		return todo.IF[[]byte](v, True, False), nil
	}
	return nil, fmt.Errorf("unable encode must bool , got: %T", a)
}

type UInteger[T uint8 | uint16 | uint32 | uint | uint64] struct{}

func (UInteger[T]) TypeFor() any                        { var t T; return t }
func (UInteger[T]) MimeDecode(data []byte) (any, error) { return UInt[T](data, 0) }
func (UInteger[T]) MimeEncode(a any) ([]byte, error)    { return FormatUInt[T](a) }

type Integer[T int8 | int16 | int32 | int | int64] struct{}

func (Integer[T]) TypeFor() any                        { var t T; return t }
func (Integer[T]) MimeDecode(data []byte) (any, error) { return Int[T](data, 0) }
func (Integer[T]) MimeEncode(a any) ([]byte, error)    { return FormatInt[T](a) }

type Float[T float32 | float64] struct{}

func (Float[T]) TypeFor() any { var t T; return t }

func (Float[T]) MimeDecode(data []byte) (any, error) {
	text := cast.B2S(data)
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return T(0), err
	}
	return T(v), nil
}

func (Float[T]) MimeEncode(a any) ([]byte, error) {
	v, ok := a.(T)
	if ok {
		data := strconv.FormatFloat(float64(v), 'f', -1, 64)
		return cast.S2B(data), nil
	}
	return nil, fmt.Errorf("unable encode must:%T got:%T", T(0), a)
}

type Time struct{}

func (Time) TypeFor() any                        { return time.Time{} }
func (Time) MimeDecode(data []byte) (any, error) { return Int[int64](data, 64) }
func (Time) MimeEncode(a any) ([]byte, error) {
	if v, ok := a.(time.Time); ok {
		return FormatInt[int64](v.Unix())
	}
	return nil, fmt.Errorf("unable encode must:%T got:%T", time.Time{}, a)
}
