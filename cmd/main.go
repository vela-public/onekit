package main

import (
	"fmt"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"path/filepath"
)

type Config struct {
	ID   int      `lua:"id"`
	Name string   `lua:"name"`
	Addr []string `lua:"addr"`
}

func Decoder(L *lua.LState) int {
	tab := L.CheckTable(1)

	cfg := &Config{}

	err := luakit.TableTo(L, tab, cfg)
	if err != nil {
		L.Push(lua.LString("decode config error: " + err.Error()))
		return 1
	}
	return 0
}

func main() {

	file := "cmd/vela.lua"

	co := lua.NewState()
	defer co.Close()
	co.SetGlobal("decode", co.NewFunction(Decoder))
	luakit.PreloadModule(co)

	if err := co.DoFile(file); err != nil {
		panic(err)
	}

	fmt.Println(filepath.Dir("/root/index/a.lua"))

}