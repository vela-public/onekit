package netkit

import (
	"github.com/vela-public/onekit/lua"
)

func newLuaNetCat(L *lua.LState) int {
	nc := newNC(L.IsString(1))
	nc.request(L.IsString(2))
	L.Push(nc)
	return 1
}

func Preload(loader lua.Preloader) {
	kv := lua.NewUserKV()
	kv.Set("ipv4", lua.NewFunction(newLuaIpv4))
	kv.Set("ipv6", lua.NewFunction(newLuaIPv6))
	kv.Set("ip", lua.NewFunction(newLuaIP))
	kv.Set("ping", lua.NewFunction(newLuaPing))
	kv.Set("cat", lua.NewFunction(newLuaNetCat))
	loader.SetGlobal("netkit", lua.NewExport("lua.netkit.export", lua.WithTable(kv)))
}
