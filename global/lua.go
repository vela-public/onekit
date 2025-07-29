package global

import "github.com/vela-public/onekit/lua"

func Preload(p lua.Preloader) {
	p.Set("var", NewVariable())
}
