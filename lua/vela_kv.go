package lua

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type ExDataKV struct {
	key   string
	value interface{}
}

type ExData []ExDataKV

func (ee *ExData) Clone() *ExData {
	sz := ee.Len()
	e2 := make(ExData, sz)
	a := *ee
	for i := 0; i < sz; i++ {
		e2[i] = a[i]
	}
	return &e2
}

func (ee *ExData) Len() int { return len(*ee) }

func (ee *ExData) Swap(i, j int) {
	a := *ee
	a[i], a[j] = a[j], a[i]
}

func (ee *ExData) Less(i, j int) bool {
	a := *ee
	return a[i].key < a[j].key
}

func (ee *ExData) Set(key string, value interface{}) {
	args := *ee

	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if key == kv.key {
			kv.value = value
			return
		}
		if kv.key == "" {
			kv.value = value
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.key = key

		kv.value = value
		*ee = args
		return
	}

	kv := ExDataKV{}
	kv.key = key
	kv.value = value
	*ee = append(args, kv)

	//排序
	sort.Sort(ee)
}

func (ee *ExData) Get(key string) interface{} {

	a := *ee
	i, j := 0, ee.Len()
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		switch strings.Compare(key, a[h].key) {
		case 0:
			return a[h]
		case 1:
			i = h + 1
		case -1:
			j = h
		}
	}

	return nil
}

func (ee *ExData) Del(key string) {
	a := *ee
	n := len(a)
	for i := 0; i < n; i++ {
		kv := &a[i]
		if kv.key == key {
			a = append(a[:i], a[:i+1]...)
			goto DONE
		}
	}

DONE:
	*ee = a
}

func (ee *ExData) Reset() {
	*ee = (*ee)[:0]
}

type exUserKV struct {
	key string
	val LValue
}

type UserKV interface {
	LValue
	Get(string) LValue
	Set(string, LValue)
	V(string) (LValue, bool)
	ForEach(func(key string, val LValue) (stop bool))
}

type userKV struct {
	data  []exUserKV
	index func(*LState, string) LValue
}

func NewUserKV() UserKV {
	return &userKV{}
}

func NewUserKVWithIndex(index func(*LState, string) LValue) UserKV {
	return &userKV{index: index}
}

func (u *userKV) Len() int {
	return len(u.data)
}

func (u *userKV) ForEach(fn func(key string, val LValue) (stop bool)) {
	n := u.Len()
	for i := 0; i < n; i++ {
		kv := &u.data[i]
		if !fn(kv.key, kv.val) {
			break
		}
	}
}

func (u *userKV) cap() int {
	return cap(u.data)
}

func (u *userKV) Get(key string) LValue {
	n := u.Len()
	for i := 0; i < n; i++ {
		kv := &u.data[i]
		if kv.key == key {
			return kv.val
		}
	}
	return LNil
}

func (u *userKV) Set(key string, val LValue) {
	n := u.Len()
	for i := 0; i < n; i++ {
		kv := &u.data[i]
		if key == kv.key {
			kv.val = val
			return
		}
	}

	c := u.cap()
	if c > n {
		u.data = u.data[:n+1]
		kv := &u.data[n]
		kv.key = key
		kv.val = val
		return
	}

	kv := exUserKV{}
	kv.key = key
	kv.val = val

	u.data = append(u.data, kv)
}

func (u *userKV) V(key string) (LValue, bool) {
	n := u.Len()
	for i := 0; i < n; i++ {
		kv := &u.data[i]
		if kv.key == key {
			return kv.val, true
		}
	}
	return nil, false
}

func (u *userKV) String() string                     { return fmt.Sprintf("function: %p", u) }
func (u *userKV) Type() LValueType                   { return LTKv }
func (u *userKV) AssertFloat64() (float64, bool)     { return 0, false }
func (u *userKV) AssertString() (string, bool)       { return "", false }
func (u *userKV) AssertFunction() (*LFunction, bool) { return nil, false }
func (u *userKV) Hijack(*CallFrameFSM) bool          { return false }
func (u *userKV) Index(L *LState, key string) LValue {
	v := u.Get(key)
	if v.Type() != LTNil {
		return v
	}

	if u.index != nil {
		return u.index(L, key)
	}
	return LNil
}

type safeUserKV struct {
	sync.RWMutex
	data  []exUserKV
	index func(*LState, string) LValue
}

func NewSafeUserKV() UserKV {
	return &safeUserKV{}
}

func NewSafeUserKVWithIndex(index func(*LState, string) LValue) UserKV {
	return &safeUserKV{
		index: index,
	}
}

func (ss *safeUserKV) Len() int {
	return len(ss.data)
}

func (ss *safeUserKV) cap() int {
	return cap(ss.data)
}

func (ss *safeUserKV) Swap(i, j int) {
	ss.data[i], ss.data[j] = ss.data[j], ss.data[i]
}

func (ss *safeUserKV) Less(i, j int) bool {
	return ss.data[i].key < ss.data[j].key
}

func (ss *safeUserKV) ForEach(fn func(key string, val LValue) (stop bool)) {
	n := ss.Len()
	for i := 0; i < n; i++ {
		kv := &ss.data[i]
		if !fn(kv.key, kv.val) {
			break
		}
	}
}

func (ss *safeUserKV) reset() {
	ss.Lock()

	n := ss.Len()
	for i := 0; i < n; i++ {
		ss.data = nil
	}
	ss.data = ss.data[:0]
	ss.Unlock()
}

func (ss *safeUserKV) Set(key string, val LValue) {
	ss.Lock()

	n := ss.Len()
	c := ss.cap()

	var newKV exUserKV
	for i := 0; i < n; i++ {
		kv := &ss.data[i]
		if key == kv.key {
			kv.val = val
			goto done
		}
		if kv.key == "" {
			kv.val = val
			goto done
		}
	}

	if c > n {
		ss.data = ss.data[:n+1]
		kv := &ss.data[n]
		kv.key = key
		kv.val = val
	}

	newKV = exUserKV{}
	newKV.key = key
	newKV.val = val
	ss.data = append(ss.data, newKV)

done:
	//排序
	sort.Sort(ss)
	ss.Unlock()
}

// 获取
func (ss *safeUserKV) Get(key string) LValue {
	ss.RLock()
	i, j := 0, ss.Len()
	val := LNil
	for i < j {
		h := int(uint(i+j) >> 1)
		switch strings.Compare(key, ss.data[h].key) {
		case 0:
			val = ss.data[h].val
			goto done
		case 1:
			i = h + 1
		case -1:
			j = h
		}
	}

done:
	ss.RUnlock()
	return val
}

func (ss *safeUserKV) V(key string) (LValue, bool) {
	ss.RLock()
	defer ss.RUnlock()
	i, j := 0, ss.Len()
	for i < j {
		h := int(uint(i+j) >> 1)
		switch strings.Compare(key, ss.data[h].key) {
		case 0:
			return ss.data[h].val, true

		case 1:
			i = h + 1
		case -1:
			j = h
		}
	}
	return nil, false
}

func (ss *safeUserKV) String() string                     { return fmt.Sprintf("function: %p", ss) }
func (ss *safeUserKV) Type() LValueType                   { return LTSkv }
func (ss *safeUserKV) AssertFloat64() (float64, bool)     { return 0, false }
func (ss *safeUserKV) AssertString() (string, bool)       { return "", false }
func (ss *safeUserKV) AssertFunction() (*LFunction, bool) { return nil, false }
func (ss *safeUserKV) Hijack(*CallFrameFSM) bool          { return false }
func (ss *safeUserKV) Index(L *LState, key string) LValue {
	v := ss.Get(key)
	if v.Type() != LTNil {
		return v
	}

	if ss.index != nil {
		return ss.index(L, key)
	}
	return LNil
}
