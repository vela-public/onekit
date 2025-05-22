package libkit

import (
	"bytes"
)

type DataKey interface {
	string | uint8 | int8 | uint16 | int16 | uint | int | uint32 | int32 | uint64 | int64 | []byte | bool | float64 | float32
}

func Equal[T DataKey](a, b T) bool {
	switch v := any(a).(type) {
	case string:
		return v == any(b).(string)
	case uint:
		return v == any(b).(uint)
	case int:
		return v == any(b).(int)
	case uint64:
		return v == any(b).(uint64)
	case int64:
		return v == any(b).(int64)
	case []byte:
		return bytes.Equal(v, any(b).([]byte))
	case bool:
		return v == any(b).(bool)
	case float64:
		return v == any(b).(float64)
	case float32:
		return v == any(b).(float32)
	default:
		return false
	}
}

type KeyVal[K DataKey, V any] struct {
	Key   K `json:"key"`
	Value V `json:"value"`
}

type DataKV[K DataKey, V any] []KeyVal[K, V]

func (d *DataKV[K, V]) Len() int {
	return len(*d)
}

func (d *DataKV[K, V]) Swap(i, j int) {
	a := *d
	a[i], a[j] = a[j], a[i]
}

func (d *DataKV[K, V]) Less(i, j int) bool {
	a := *d
	k1 := a[i].Key
	k2 := a[j].Key

	switch v := any(k1).(type) {
	case string:
		return v < any(k2).(string)
	case int8:
		return v < any(k2).(int8)
	case uint8:
		return v < any(k2).(uint8)
	case int16:
		return v < any(k2).(int16)
	case uint16:
		return v < any(k2).(uint16)
	case int32:
		return v < any(k2).(int32)
	case uint32:
		return v < any(k2).(uint32)
	case int:
		return v < any(k2).(int)
	case int64:
		return v < any(k2).(int64)
	case uint64:
		return v < any(k2).(uint64)
	case []byte:
		return bytes.Compare(v, any(k2).([]byte)) < 0
	default:
		return false
	}
}

func (d *DataKV[K, V]) Set(key K, value V) {
	args := *d

	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if Equal(key, kv.Key) {
			kv.Value = value
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.Key = key

		kv.Value = value
		*d = args
		return
	}

	kv := KeyVal[K, V]{
		Key:   key,
		Value: value,
	}
	*d = append(args, kv)
}

func (d *DataKV[K, V]) Get(key K) (v V) {
	sz := d.Len()
	for i := 0; i < sz; i++ {
		kv := &(*d)[i]
		if Equal(key, kv.Key) {
			return kv.Value
		}
	}
	return
}

func (d *DataKV[K, V]) Range(f func(key K, value V)) {
	for _, kv := range *d {
		f(kv.Key, kv.Value)
	}
}

func (d *DataKV[K, V]) Del(key K) {
	a := *d
	n := len(a)
	for i := 0; i < n; i++ {
		kv := &a[i]
		if Equal(kv.Key, key) {
			a = append(a[:i], a[i+1:]...)
			goto DONE
		}
	}

DONE:
	*d = a
}

func NewDataKV[K DataKey, V any]() *DataKV[K, V] {
	return new(DataKV[K, V])
}
