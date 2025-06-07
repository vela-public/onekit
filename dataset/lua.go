package dataset

import "github.com/vela-public/onekit/lua"

/*
	local dataset = vela.dataset.create(int32(4))
	local cdata = dataset.acquire()
	cdata[0] = 1
	cdata[1] = 2
	cdata[2] = 3
	cdata[3] = 4
	dataset.release(cdata)

	local dataset = vela.dataset.create(packet("master:int", "slave:int", "black:text(100)"))
	local cdata = dataset.acquire()
	cdata.master = 200
	cdata.slave = 100
	cdata.black = "hello world"
	dataset.release(cdata)

*/

func Preload(p lua.Preloader) {
	kv := lua.NewUserKV()
	kv.Set("int32", lua.NewFunction(NewInt32L))
	p.Set("dataset", lua.NewExport("lua.dataset.export", lua.WithTable(kv)))
}
