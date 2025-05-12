package render

import (
	"github.com/valyala/fasttemplate"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/lua"
	"io"
	"strings"
	"time"
)

type ReadFunc func() (tex string, update bool)

type Render struct {
	Left     string
	Right    string
	Reader   ReadFunc
	Template *fasttemplate.Template
	need     bool
}

func (r *Render) builtin(tag string) []byte {
	switch tag {
	case "@today":
		return cast.S2B(time.Now().Format("2006-01-02"))
	case "@year":
		return cast.S2B(time.Now().Format("2006"))
	case "@month":
		return cast.S2B(time.Now().Format("01"))
	case "@day":
		return cast.S2B(time.Now().Format("02"))
	case "@hour":
		return cast.S2B(time.Now().Format("15"))
	}
	return nil
}

func (r *Render) PrepareText() {
	text, _ := r.Reader()
	if strings.Index(text, r.Left) == -1 {
		r.need = false
		return
	}
	r.need = true
	r.Template = fasttemplate.New(text, r.Left, r.Right)
}

func Extract(v any, env *Env) func(string) (string, bool) {
	switch vt := v.(type) {
	case nil:
		return func(name string) (string, bool) {
			return "", false
		}

	case map[string]any:
		return func(name string) (string, bool) {
			val, ok := vt[name]
			if !ok {
				return "", false
			}
			return cast.ToString(val), true
		}

	case map[string]string:
		return func(name string) (string, bool) {
			val, ok := vt[name]
			if !ok {
				return "", false
			}
			return val, true
		}
	case lua.IndexType:
		return func(name string) (string, bool) {
			vv := vt.Index(env.LState, name)
			if vv == nil || vv.Type() == lua.LTNil {
				return "", false
			}
			return vv.String(), true
		}
	case lua.MetaType:
		return func(name string) (string, bool) {
			vv := vt.Meta(env.LState, lua.S2L(name))
			if vv == nil || vv.Type() == lua.LTNil {
				return "", false
			}
			return vv.String(), true
		}
	case lua.MetaTableType:
		return func(name string) (string, bool) {
			vv := vt.MetaTable(env.LState, name)
			if vv == nil || vv.Type() == lua.LTNil {
				return "", false
			}
			return vv.String(), true
		}
	case lua.FieldType:
		return func(name string) (string, bool) {
			vv := vt.Field(name)
			if vv == "" {
				return "", false
			}
			return vv, true
		}

	case *lua.LTable:
		return func(name string) (string, bool) {
			vv := vt.RawGet(lua.S2L(name))
			if vv == nil || vv.Type() == lua.LTNil {
				return "", false
			}
			return vv.String(), true
		}
	}
	return func(name string) (string, bool) {
		return "", false
	}
}

func (r *Render) Render(v any, env *Env) string {
	if !r.need {
		text, _ := r.Reader()
		return text
	}
	if env == nil {
		env = &Env{}
	}

	content, change := r.Reader()
	if change || r.Template == nil {
		r.Template = fasttemplate.New(content, r.Left, r.Right)
	}

	extract := Extract(v, env)
	result := r.Template.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		text, ok := extract(tag)
		if ok {
			return w.Write(cast.S2B(text))
		}
		return w.Write(r.builtin(tag))
	})
	return result
}

func Tag(left, right string) func(*Render) {
	return func(r *Render) {
		r.Left = left
		r.Right = right
	}
}

func NewRender(reader ReadFunc, options ...func(*Render)) *Render {
	r := &Render{Reader: reader, Left: "${", Right: "}"}
	for _, option := range options {
		option(r)
	}
	return r
}

func Text(text string, options ...func(*Render)) *Render {
	r := NewRender(func() (string, bool) {
		return text, false
	})

	for _, option := range options {
		option(r)
	}

	r.PrepareText()
	return r
}
