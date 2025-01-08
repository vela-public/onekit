package sax

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
)

func Noop(string) string {
	return ""
}

type Extractor struct {
	co       *lua.LState
	object   interface{}
	function func(string) string
}

func (e *Extractor) parse() error {

	switch entry := e.object.(type) {
	case lua.IndexType:
		e.function = func(key string) string {
			return entry.Index(e.co, key).String()
		}
	case lua.MetaType:
		e.function = func(key string) string {
			return entry.Meta(e.co, lua.S2L(key)).String()
		}

	case lua.MetaTableType:
		e.function = func(key string) string {
			return entry.MetaTable(e.co, key).String()
		}

	case map[string]string:
		e.function = func(key string) string {
			return entry[key]
		}

	case map[string]interface{}:
		e.function = func(key string) string {
			return cast.ToString(entry[key])
		}

	case string:
		e.function = Kind(entry)

	case []byte:
		e.function = Kind(cast.B2S(entry))

	default:
		return fmt.Errorf("extractor parse fail")
	}

	return nil
}

func (e *Extractor) Peek(name string) string {
	if e.function != nil {
		return e.function(name)
	}
	return ""
}

func NewExtractor(v interface{}, co *lua.LState) (*Extractor, error) {
	e := &Extractor{object: v, co: co}
	return e, e.parse()
}
