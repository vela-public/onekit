package ruleset

import (
	"encoding/json"
	"github.com/vela-public/onekit/libkit"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/pipe"
	"github.com/vela-public/onekit/treekit"
	"strings"
)

type Rule struct {
	Key   string
	Chain *pipe.Chain
}

func (r *Rule) Nil() bool {
	if r.Key == "" || r.Chain == nil {
		return true
	}

	return false
}

type RuleSet struct {
	L    *lua.LState
	name string
	Data []*Rule
}

func (rs *RuleSet) Name() string {
	return rs.name
}

func (rs *RuleSet) Startup(env *treekit.Env) error {
	return nil
}

func (rs *RuleSet) Close() error {
	return nil
}

func (rs *RuleSet) Metadata() libkit.DataKV[string, any] {
	return libkit.DataKV[string, any]{}
}

func (rs *RuleSet) Text() []byte {
	text, _ := json.Marshal(rs)
	return text
}

func (rs *RuleSet) String() string                         { return lua.B2S(rs.Text()) }
func (rs *RuleSet) Type() lua.LValueType                   { return lua.LTObject }
func (rs *RuleSet) AssertFloat64() (float64, bool)         { return float64(len(rs.Data)), true }
func (rs *RuleSet) AssertString() (string, bool)           { return "", false }
func (rs *RuleSet) AssertFunction() (*lua.LFunction, bool) { return nil, false }
func (rs *RuleSet) Hijack(fsm *lua.CallFrameFSM) bool      { return false }

func (rs *RuleSet) Exec(v any, names ...string) {
	sz := len(names)
	for i := 0; i < sz; i++ {
		name := names[i]
		idx := rs.Have(name)
		if idx == -1 {
			continue
		}
		r := rs.At(idx)
		if r.Nil() || r.Chain.Len() == 0 {
			return
		}
		r.Chain.Invoke(v)
	}

}
func (rs *RuleSet) Use(rules ...string) lua.Invoker {
	return func(v any) error {
		rs.Exec(v, rules...)
		return nil
	}
}

func (rs *RuleSet) UseTag(prefix string) lua.Invoker {
	return func(v any) error {
		var tag string
		switch vt := v.(type) {
		case lua.IndexType:
			tag = vt.Index(rs.L, "tag").String()
		case lua.IndexOfType:
			tag = vt.IndexOf(rs.L, "tag").String()
		case lua.MetaType:
			tag = vt.Meta(rs.L, lua.LString("tag")).String()
		case lua.MetaTableType:
			tag = vt.MetaTable(rs.L, "tag").String()
		}

		if len(tag) == 0 {
			return nil
		}
		if prefix != "" {
			tag = prefix + tag
		}

		rs.Exec(v, tag)
		return nil
	}
}

func (rs *RuleSet) UseL(L *lua.LState) int {
	rules := lua.Unpack[string](L)
	L.Push(rs.Use(rules...))
	return 1
}

func (rs *RuleSet) UseTagL(L *lua.LState) int {
	prefix := L.IsString(1)
	L.Push(rs.UseTag(prefix))
	return 1
}

func (rs *RuleSet) DynamicL(L *lua.LState) int {
	text := L.CheckString(1) // public.v_black > 10 then deny
	//todo

	/*
		deny = function()
			line.incr('')
		end

		function(lock , deny , line)
			if public.v_black > 10 then deny(); return end
		    if public.v_black > 10 then lock(100); return end


	*/
	L.DoString(text)
	rs.Add(&Rule{
		Key:   "hash",
		Chain: nil,
	})

	return 0
}

func (rs *RuleSet) Index(L *lua.LState, key string) lua.LValue {
	switch key {
	case "use":
		return lua.NewFunction(rs.UseL)
	case "tag":
		return lua.NewFunction(rs.UseTagL)
	case "dynamic":
		return lua.NewFunction(rs.DynamicL)
	}

	if strings.HasPrefix(key, "use_") {
		name := strings.TrimPrefix(key, "use_")
		if len(name) == 0 {
			return lua.LNil
		}
		return rs.Use(name)
	}
	return lua.LNil
}

func (rs *RuleSet) NewIndex(L *lua.LState, key string, val lua.LValue) {
	if strings.HasPrefix(key, "do_") {
		L.RaiseError("not allow do_ prefix")
		return
	}

	idx := rs.Have(key)
	if idx == -1 {
		rs.Add(&Rule{
			Key:   key,
			Chain: pipe.LValue(val, pipe.LState(L)),
		})
		return
	}

	rs.At(idx).Chain.Merge(pipe.LValue(val, pipe.LState(L)))
}

func (rs *RuleSet) Len() int {
	return len(rs.Data)
}

func (rs *RuleSet) At(i int) *Rule {
	sz := rs.Len()
	if i < 0 || i >= sz {
		return &Rule{}
	}
	return rs.Data[i]
}

func (rs *RuleSet) Have(key string) int {
	sz := rs.Len()
	if sz == 0 {
		return -1
	}

	for i := 0; i < sz; i++ {
		r := rs.At(i)
		if !r.Nil() && r.Key == key {
			return i
		}
	}
	return -1
}
func (rs *RuleSet) Add(rule *Rule) {
	if rs.Have(rule.Key) != -1 {
		return
	}
	rs.Data = append(rs.Data, rule)
}
