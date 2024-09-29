package pipe

import "github.com/vela-public/onekit/lua"

func Seek(n int) func(*Chains) {
	return func(px *Chains) {
		if n < 0 {
			return
		}
		px.seek = n
	}
}

func Env(env Environment) func(*Chains) {
	return func(px *Chains) {
		px.xEnv = env
	}
}

func LState(L *lua.LState) func(*Chains) {
	return func(c *Chains) {
		c.vm = L
	}

}
