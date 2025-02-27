package strkit

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/netkit"
	"path/filepath"
	"strings"
)

func Noop(string) string {
	return ""
}

func Kind(raw string) func(string) string {

	size := len(raw)

	return func(key string) string { // * , ext , ipv4, ipv6 , [1,3]
		switch key {
		case "*":
			return raw
		case "ext":
			return filepath.Ext(raw)
		case "ipv4":
			return cast.ToString(netkit.Ipv4(raw))
		case "ipv6":
			return cast.ToString(netkit.Ipv6(raw))
		case "ip":
			return cast.ToString(netkit.Ipv4(raw) || netkit.Ipv6(raw))
		}

		n := len(key)
		if n < 3 {
			return raw
		}

		if key[0] != '[' {
			return raw
		}

		if key[n-1] != ']' {
			return raw
		}

		idx := strings.Index(key, ":")
		if idx < 0 {
			offset, err := cast.ToIntE(key[1 : n-1])
			if err != nil {
				return raw
			}

			if offset >= 1 && offset <= len(raw) {
				return string(raw[offset-1])
			}

			return raw
		}

		s := cast.ToInt(key[1:idx])
		e := cast.ToInt(key[idx+1 : n-1])
		if s > size {
			return ""
		}

		if e == 0 || e > size {
			return raw[s:]
		}

		if s > e {
			return ""
		}

		return raw[s:e]
	}
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
