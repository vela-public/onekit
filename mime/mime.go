package mime

import (
	"reflect"
	"sync"
)

var (
	mt = &Table{
		tab: make(map[string]TypeOf),
	}
)

type Table struct {
	mux sync.RWMutex
	tab map[string]TypeOf
}

func Encode(v interface{}) ([]byte, string, error) {
	name := Name(v)
	mt.mux.RLock()
	defer mt.mux.RUnlock()

	t, ok := mt.tab[name]
	if !ok {
		return nil, name, NotFound
	}

	data, err := t.MimeEncode(v)
	if err == nil {
		return data, name, nil
	}
	return nil, name, err
}

func Decode(name string, data []byte) (interface{}, error) {
	mt.mux.RLock()
	defer mt.mux.RUnlock()

	t, ok := mt.tab[name]
	if !ok {
		return nil, NotFound
	}
	return t.MimeDecode(data)
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

func Register(t TypeOf) {
	mt.mux.Lock()
	defer mt.mux.Unlock()

	name := Name(t.TypeFor())
	_, ok := mt.tab[name]
	if ok {
		panic("duplicate mime " + name)
		return
	}
	mt.tab[name] = t
}

func TypeFor[T any]() {
	mt.mux.Lock()
	defer mt.mux.Unlock()

	var vt T
	var of TypeOf
	name := Name(vt)

	_, ok := mt.tab[name]
	if ok {
		return
	}

	of, ok = any(vt).(TypeOf)
	if ok {
		mt.tab[name] = of
		return
	}

	mt.tab[name] = Unknown[T]{}
}

func init() {
	Register(Nil{})   //nil
	Register(Text{})  //string
	Register(Bytes{}) //[]byte
	Register(Bool{})  //bool
	Register(UInteger[uint8]{})
	Register(UInteger[uint16]{})
	Register(UInteger[uint32]{})
	Register(UInteger[uint]{})
	Register(UInteger[uint64]{})
	Register(Integer[int8]{})
	Register(Integer[int16]{})
	Register(Integer[int32]{})
	Register(Integer[int]{})
	Register(Integer[int64]{})
	Register(Float[float32]{})
	Register(Float[float64]{})
	Register(Time{})
}
