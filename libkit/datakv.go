package libkit

import (
	"bytes"
)

type DataKey interface {
	string | uint | int | uint64 | int64 | []byte | bool | float64 | float32
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
	key K
	val V
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
	k1 := a[i].key
	k2 := a[j].key

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
		if Equal(key, kv.key) {
			kv.val = value
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.key = key

		kv.val = value
		*d = args
		return
	}

	kv := KeyVal[K, V]{
		key: key,
		val: value,
	}
	*d = append(args, kv)
}

func (d *DataKV[K, V]) Get(key K) (v V) {
	sz := d.Len()
	for i := 0; i < sz; i++ {
		kv := &(*d)[i]
		if Equal(key, kv.key) {
			return kv.val
		}
	}
	return
}

func (d *DataKV[K, V]) Del(key K) {
	a := *d
	n := len(a)
	for i := 0; i < n; i++ {
		kv := &a[i]
		if Equal(kv.key, key) {
			a = append(a[:i], a[:i+1]...)
			goto DONE
		}
	}

DONE:
	*d = a
}
