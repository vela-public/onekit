package cond

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/errkit"
	"github.com/vela-public/onekit/lua"
	"time"
)

type OptionFunc func(*option)

type option struct {
	seek      int
	value     interface{}
	logic     CndMode
	field     Lookup
	errs      *errkit.JoinError
	compare   func(string, string, Method) bool
	co        *lua.LState
	partition []int
	payload   func(int, string)
}

func Seek(i int) OptionFunc {
	return func(o *option) {
		o.seek = i
	}
}

func Mode(v CndMode) OptionFunc {
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
	case Lookup:
		opt.field = item
		return true
	case CompareEx:
		opt.compare = item.Compare

	case string:
		opt.field = String(item)
		return true

	case []byte:
		opt.field = String(string(item))
		return true
	case map[string]float32:
		opt.field = func(key string) string {
			return cast.ToString(item[key])
		}
		return true
	case map[string]float64:
		opt.field = func(key string) string {
			return cast.ToString(item[key])
		}
		return true
	case map[string]int64:
		opt.field = func(key string) string {
			return cast.ToString(item[key])
		}
		return true
	case map[string]uint64:
		opt.field = func(key string) string {
			return cast.ToString(item[key])
		}
		return true
	case map[string]time.Time:
		opt.field = func(key string) string {
			return cast.ToString(item[key])
		}
		return true

	case map[string]int32:
		opt.field = func(key string) string {
			return cast.ToString(item[key])
		}
		return true

	case map[string]int:
		opt.field = func(key string) string {
			return cast.ToString(item[key])
		}
		return true
	case map[string]string:
		opt.field = func(key string) string {
			return item[key]
		}
		return true
	case map[string]bool:
		opt.field = func(key string) string {
			if item[key] {
				return "true"
			}
			return "false"
		}
		return true
	case map[string]interface{}:
		opt.field = func(key string) string {
			return cast.ToString(item[key])
		}
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
	case lua.PackType:
		return opt.NewPeek(item.Unpack())
	}

	return false
}
