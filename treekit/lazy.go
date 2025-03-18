package treekit

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"reflect"
	"strconv"
)

type BuildOf[T any, K any] interface {
	Build(*K) (*T, error)
}
type ReBuildOf[T any, K any] interface {
	ReBuild(*K, *T) error
}

type NameOf interface {
	Name() string
}

type LazyProcess[T any, K any] struct {
	co     *lua.LState
	conf   *K
	data   *T
	name   string
	typeof string

	handler struct {
		Error func(error)
		Panic func(error)
	}

	private struct {
		Ready  bool
		ExData *Process
		Info   error
	}
}

func Lazy[T any, K any](L *lua.LState, idx int) *LazyProcess[T, K] {
	l := &LazyProcess[T, K]{
		co:     L,
		conf:   new(K),
		typeof: reflect.TypeOf((*T)(nil)).String(),
	}

	l.SetPanicHandler(L.PanicErr)
	l.SetErrorHandler(L.PanicErr)
	l.MustBe((*T)(nil))
	l.TableTo(L, idx)
	l.define(L, l.Name(), l.typeof)
	return l
}

func LazyCNF[T any, K any](L *lua.LState, conf *K) *LazyProcess[T, K] {
	l := &LazyProcess[T, K]{
		co:     L,
		conf:   conf,
		typeof: reflect.TypeOf((*T)(nil)).String(),
	}
	l.SetPanicHandler(L.PanicErr)
	l.SetErrorHandler(L.PanicErr)
	l.MustBe((*T)(nil))
	l.define(L, l.Name(), l.typeof)
	return l
}

func (l *LazyProcess[T, K]) TableTo(L *lua.LState, idx int) *LazyProcess[T, K] {
	tab := L.CheckTable(idx)

	if l.conf == nil {
		l.conf = new(K)
	}

	err := luakit.TableTo(L, tab, l.conf)
	if err != nil {
		L.PanicErr(err)
	}
	return l
}

func (l *LazyProcess[T, K]) Name() string {
	if len(l.name) > 0 {
		goto done
	}

	switch cnf := any(l.conf).(type) {
	case NameOf:
		l.name = cnf.Name()
		goto done
	case *string:
		return *cnf
	case *[]byte:
		return string(*cnf)
	case *int:
		return strconv.Itoa(*cnf)
	case *int64:
		return strconv.FormatInt(*cnf, 10)
	case *int32:
		return strconv.FormatInt(int64(*cnf), 10)
	case *int16:
		return strconv.FormatInt(int64(*cnf), 10)
	case *int8:
		return strconv.FormatInt(int64(*cnf), 10)
	case *uint:
		return strconv.FormatUint(uint64(*cnf), 10)
	case *uint64:
		return strconv.FormatUint(*cnf, 10)
	case *uint32:
		return strconv.FormatUint(uint64(*cnf), 10)
	case *uint16:
		return strconv.FormatUint(uint64(*cnf), 10)
	case *uint8:
		return strconv.FormatUint(uint64(*cnf), 10)
	case *float32:
		return strconv.FormatFloat(float64(*cnf), 'f', -1, 32)
	case *float64:
		return strconv.FormatFloat(*cnf, 'f', -1, 64)
	default:
		if f, ok := lua.NewReflect[*K](l.conf).IndexOf(l.co, "name").(lua.LString); ok {
			l.name = f.String()
			goto done
		}
		l.co.RaiseError("%s name not found", l.conf)
	}

done:
	return l.name
}

func (l *LazyProcess[T, K]) SetErrorHandler(fn func(error)) {
	l.handler.Error = fn
}

func (l *LazyProcess[T, K]) SetPanicHandler(fn func(error)) {
	l.handler.Panic = fn
}

func (l *LazyProcess[T, K]) Start() {
	pro := l.Unwrap()
	Start(l.co, pro.data, func(err error) {
		l.onError(err)
	})
}

func (l *LazyProcess[T, K]) onError(err error) {
	if l.handler.Error != nil {
		l.handler.Error(err)
		return
	}

	l.co.PanicErr(err)
}

func (l *LazyProcess[T, K]) onPanic(err error) {
	if l.handler.Panic != nil {
		l.handler.Panic(err)
		return
	}
	l.co.PanicErr(err)
}

func (l *LazyProcess[T, K]) MustBe(v any) ProcessType {
	pt, ok := v.(ProcessType)
	if !ok {
		l.handler.Panic(fmt.Errorf("mismatch process type [%T] Got %T", (*T)(nil), v))
	}
	return pt
}

func (l *LazyProcess[T, K]) Ready() bool {
	return l.private.Ready
}
func (l *LazyProcess[T, K]) readyTo(v bool) {
	l.private.Ready = v
}

func (l *LazyProcess[T, K]) Build(fn func(*K) *T) {
	if l.Ready() {
		return
	}

	if pro := l.Unwrap(); pro.Nil() {
		dat := fn(l.conf)

		if dat == nil {
			l.co.RaiseError("build %s process data is nil", l.typeof)
		}
		pro.data = any(dat).(ProcessType)
		l.readyTo(true)
	}
}

func (l *LazyProcess[T, K]) Upsert(fn func(*K) *T) {
	if l.Ready() {
		return
	}

	pro := l.Unwrap()
	if pro.Nil() {
		dat := fn(l.conf)
		if dat == nil {
			l.co.RaiseError("build %s process data is nil", l.typeof)
			return
		}
		pro.data = any(dat).(ProcessType)
		l.readyTo(true)
		return
	}

	if e := pro.Close(); e != nil {
		l.co.RaiseError("%v", e)
		return
	}

	pro.data = any(fn(l.conf)).(ProcessType)
	l.readyTo(true)
}

func (l *LazyProcess[T, K]) Rebuild(fn func(*K, *T)) {
	if l.Ready() {
		return
	}

	if l.Unwrap().Nil() {
		return
	}
	fn(l.conf, l.Data())
	l.readyTo(true)
}

func (l *LazyProcess[T, K]) Close() error {
	return l.Unwrap().Close()
}

func (l *LazyProcess[T, K]) UnwrapErr() error {
	return l.private.Info
}

func (l *LazyProcess[T, K]) Unwrap() *Process {
	return l.private.ExData
}

func (l *LazyProcess[T, K]) Set(data ProcessType) {
	l.private.ExData.Set(data)
}

func (l *LazyProcess[T, K]) Data() *T {
	return any(l.Unwrap().data).(*T)
}

func (l *LazyProcess[T, K]) buildOf() func(*K) (*T, error) {
	v := new(T)

	bo, ok := any(v).(BuildOf[T, K])
	if ok {
		return bo.Build
	}

	return nil
}

func (l *LazyProcess[T, K]) rebuildOf() func(*K, *T) error {
	bo, ok := any(l.Unwrap().data).(ReBuildOf[T, K])
	if ok {
		return bo.ReBuild
	}

	return nil

}

func (l *LazyProcess[T, K]) define(L *lua.LState, name string, typeof string) {
	ref := func(v *Process) {
		l.private.ExData = v
		if build := l.buildOf(); build != nil && v.Nil() {
			dat, err := build(l.conf)
			if err != nil {
				l.co.RaiseError("%v", err)
			}
			v.data = any(dat).(ProcessType)
			v.typeof = l.typeof
			return
		}

		if rebuild := l.rebuildOf(); rebuild != nil && !v.Nil() {
			err := rebuild(l.conf, l.Data())
			if err != nil {
				l.co.RaiseError("%v", err)
			}
			return
		}
	}

	exdata := L.Exdata()
	switch dat := exdata.(type) {
	case *MicroService:
		ref(dat.Create(L, name, typeof))
	case *Task:
		ref(dat.Create(L, name, typeof))
	default:
		L.RaiseError("lua.exdata must *MicroService or *TaskTree got:%T", exdata)
	}
}
