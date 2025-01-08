package cond

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
)

type OptionFunc func(*option)

type option struct {
	unary     bool
	seek      int
	value     interface{}
	logic     Logic
	field     FuncType
	errs      *errkit.JoinError
	compare   func(string, string, Method) bool
	co        *lua.LState
	partition []int
	payload   func(int, string)
}

func WithUnary(v bool) OptionFunc {
	return func(o *option) {
		o.unary = v
	}
}

func Seek(i int) OptionFunc {
	return func(o *option) {
		o.seek = i
	}
}

func WithLogic(v Logic) OptionFunc {
	return func(o *option) {
		o.logic = v
	}
}

func Partition(v []int) OptionFunc {
	return func(o *option) {
		o.partition = v
	}
}

func Payload(fn func(int, string)) func(*option) {
	return func(o *option) {
		o.payload = fn
	}
}

func LState(co *lua.LState) func(*option) {
	return func(ov *option) {
		ov.co = co
	}
}

func (opt *option) Pay(i int, v string) {
	if opt.payload == nil {
		return
	}
	opt.payload(i, v)
}

func (opt *option) NewPeek(v interface{}) bool {
	switch item := v.(type) {
	case FuncType:
		opt.field = item
		return true

	case FieldType:
		opt.field = item.Field
		return true

	case CompareEx:
		opt.compare = item.Compare

	case string:
		opt.field = String(item)
		return true

	case []byte:
		opt.field = String(string(item))
		return true

	case func() string:
		opt.field = func(string) string {
			return item()
		}
		return true
	case lua.IndexType:
		opt.field = func(key string) string {
			return item.Index(opt.co, key).String()
		}
		return true

	case lua.MetaType:
		opt.field = func(key string) string {
			return item.Meta(opt.co, lua.S2L(key)).String()
		}
		return true

	case lua.MetaTableType:
		opt.field = func(key string) string {
			return item.MetaTable(opt.co, key).String()
		}
		return true

	case *lua.LTable:
		opt.field = func(key string) string {
			return item.RawGetString(key).String()
		}
		return true
	case lua.Getter:
		opt.field = func(key string) string {
			return cast.ToString(item.Getter(key))
		}
		return true
	case fmt.Stringer:
		opt.field = String(item.String())
		return true
	case lua.WrapType:
		return opt.NewPeek(item.UnwrapData())
	}

	return false
}
