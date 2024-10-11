package main

import (
	"bytes"
	"fmt"
	"github.com/vela-public/onekit/bucket"
	"github.com/vela-public/onekit/lua"
	"github.com/vela-public/onekit/luakit"
	"github.com/vela-public/onekit/mime"
	"time"
)

type C2 struct {
	Key string `lua:"key"`
	Pem string `lua:"pem"`
}

type Config struct {
	ID   int      `lua:"id"`
	Name string   `lua:"name"`
	Addr []string `lua:"addr"`
	C2   *C2      `lua:"c2"`
}

type App struct {
	A, B, C int
}

func (a App) TypeFor() any { return App{} }

func (a App) MimeDecode(data []byte) (any, error) {
	s := bytes.Split(data, []byte(","))
	if len(s) != 3 {
		return nil, fmt.Errorf("data format error")
	}
	var err error

	var v App
	v.A, err = mime.Int[int](s[0], 32)
	v.B, err = mime.Int[int](s[1], 32)
	v.C, err = mime.Int[int](s[2], 32)
	return v, err
}

func (a App) MimeEncode(data any) ([]byte, error) {
	v, ok := data.(App)
	if ok {
		return []byte(fmt.Sprintf("%d,%d,%d", v.A, v.B, v.C)), nil
	}
	return nil, fmt.Errorf("data format error")
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

	db, err := bucket.Open("one.db", 0600, bucket.Default)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	type app struct{ A, B, C int }
	bkt := bucket.Pack[*app](db, "20241010")
	_ = bucket.Pack[*app](db, "20241010")
	_ = bkt.Set("k", &app{10, 100, 500}, 1000)
	fmt.Println(bkt.Get("k").Value())
	time.Sleep(time.Millisecond * 300)
	fmt.Println(bkt.Get("k").Value())
	time.Sleep(time.Millisecond * 300)
	time.Sleep(time.Millisecond * 300)
	fmt.Println(bkt.Get("k").Value())
	time.Sleep(time.Millisecond * 300)
	fmt.Println(bkt.Get("k").Value())
	//fmt.Println(filepath.Dir("/root/index/a.lua"))
	//println(mime.Name(NewA()))
	//println(NewA() == nil)
	//bkt := bucket.Pack[int](db, "2024.10.10.nil")

	//bkt := bucket.Pack[int](db, "20241010.int")
	//fmt.Println(bkt.Set("k", 10, 1000))
	//fmt.Println(bkt.Get("k").Value())
	//time.Sleep(time.Millisecond * 300)
	//fmt.Println(bkt.Get("k").Value())
	//time.Sleep(time.Millisecond * 300)
	//fmt.Println(bkt.Get("k").Value())
	//time.Sleep(time.Millisecond * 300)
	//fmt.Println(bkt.Get("k").Value())
	//time.Sleep(time.Millisecond * 300)
	//fmt.Println(bkt.Get("k").Value())
}
