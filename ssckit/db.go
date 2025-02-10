package ssckit

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/layer"
	"github.com/vela-public/onekit/lua"
	"go.etcd.io/bbolt"
	"path/filepath"
	"sync/atomic"
	"time"
)

type Database struct {
	name string
	flag struct {
		Error   error
		Compact uint32
	}
	dir     string
	opt     *bbolt.Options
	dbless  *bbolt.DB
	OnError func(string, ...any)
}

func (db *Database) UnwrapErr() error {
	return db.flag.Error
}

func (db *Database) Compacting() bool {
	return atomic.AddUint32(&db.flag.Compact, 1) >= 1
}

func (db *Database) UnCompact() {
	atomic.StoreUint32(&db.flag.Compact, 0)
}
func (db *Database) Compact() {
	if db.Compacting() {
		return
	}
	defer db.UnCompact()

	path := filepath.Join(db.dir, fmt.Sprintf(".%s-%d.db", db.name, time.Now().Unix()))
	com, err := bbolt.Open(path, 0600, db.opt)
	if err != nil {
		db.OnError("compact new %s fail %v", path, err)
		return
	}

	err = bbolt.Compact(com, db.dbless, 0)
	if err != nil {
		db.OnError("compact %s fail %v", path, err)
		return
	}
	db.dbless = com
}

func (db *Database) walk() string {
	pattern := db.dir + fmt.Sprintf("/.%s*.db", db.name)

	ms, err := filepath.Glob(pattern)
	if err != nil {
		return filepath.Join(db.dir, fmt.Sprintf(".%s.db", db.name))
	}

	n := len(ms)
	if n == 0 {
		return filepath.Join(db.dir, fmt.Sprintf(".%s.db", db.name))
	}

	FileTime := func(file string) int64 {
		base := filepath.Base(file)
		size := len(base)
		if size == 7 {
			return 0
		}

		if size != 18 {
			return -1
		}

		tv := base[5:15]
		return cast.ToInt64(tv)
	}

	fsm := new(struct {
		Max    int64
		Latest string
	})

	for i := 0; i < n; i++ {
		file := ms[i]
		m := FileTime(file)
		if m >= fsm.Max {
			fsm.Max = m
			fsm.Latest = file
		}
	}

	if fsm.Latest != "" {
		return fsm.Latest
	}

	return filepath.Join(db.dir, fmt.Sprintf(".%s.db", db.name))
}

func (db *Database) Open() {
	path := db.walk()
	dat, err := bbolt.Open(path, 0600, db.opt)
	if err != nil {
		db.flag.Error = err
		db.OnError("open %s fail %v", path, err)
		return
	}
	db.dbless = dat
}

func (db *Database) HttpView(ctx *fasthttp.RequestCtx) error {
	if db.dbless == nil {
		return fmt.Errorf("%s db not found err:%v", db.name, db.UnwrapErr())
	}

	info := new(struct {
		Num   int                          `json:"num"`
		Path  string                       `json:"path"`
		Size  int64                        `json:"size"`
		State bbolt.TxStats                `json:"state"`
		Bkt   map[string]bbolt.BucketStats `json:"bkt"`
	})

	info.Bkt = make(map[string]bbolt.BucketStats)

	err := db.dbless.View(func(tx *bbolt.Tx) error {
		info.Size = tx.Size()
		info.State = tx.Stats()
		info.Path = db.dbless.Path()
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			info.Bkt[string(name)] = b.Stats()
			info.Num++
			return nil
		})
	})

	if err != nil {
		return err
	}
	text, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = ctx.Write(text)
	return err

}

func (db *Database) HttpCompact(ctx *fasthttp.RequestCtx) error {
	if db.dbless == nil {
		return fmt.Errorf("%s db not found err:%v", db.name, db.UnwrapErr())
	}

	db.Compact()

	return db.HttpView(ctx)
}

func (db *Database) Define(r layer.RouterType) {
	_ = r.GET("/api/v1/agent/"+db.name+"/compact", r.Then(db.HttpCompact))
	_ = r.GET("/api/v1/agent/"+db.name+"/info", r.Then(db.HttpView))
}

func (db *Database) Preload(p lua.Preloader) {
}
