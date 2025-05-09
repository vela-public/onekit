package ruleset

import (
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/treekit"
)

func NewRulesetL(L *lua.LState) int {
	name := L.CheckString(1)
	pro := treekit.LazyCNF[RuleSet, string](L, &name)

	pro.Build(func(_ *string) *RuleSet {
		return &RuleSet{
			L:    L,
			name: name,
		}
	})

	pro.Rebuild(func(_ *string, r *RuleSet) {
		r.name = name
		r.Data = nil
		r.L = L
	})

	treekit.Start(L, pro.Data(), L.PanicErr)
	L.Push(pro.Unwrap())
	return 1
}

func Preload(p lua.Preloader) {
	p.Set("ruleset", lua.NewExport("lua.ruleset.export", lua.WithFunc(NewRulesetL)))
}
