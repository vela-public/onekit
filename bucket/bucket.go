package bucket

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/mime"
	"go.etcd.io/bbolt"
	"os"
	"sync"
)

var Default = bbolt.DefaultOptions

type Tx = bbolt.Tx

type Bucket[T any] struct {
	once   sync.Once
	db     *bbolt.DB
	chains [][]byte
}

func Pack[T any](db *bbolt.DB, names ...string) *Bucket[T] {
	var chains [][]byte
	for _, name := range names {
		chains = append(chains, cast.S2B(name))
	}

	mime.TypeFor[T]()
	return &Bucket[T]{
		db:     db,
		chains: chains,
	}
}

func To[T, U any](bkt *Bucket[T]) *Bucket[U] {
	return &Bucket[U]{
		db:     bkt.db,
		chains: bkt.chains,
	}
}

func Tx2B(tx *Tx, name []byte, readonly bool) (*bbolt.Bucket, error) {
	if readonly {
		bkt := tx.Bucket(name)
		if bkt == nil {
			return nil, fmt.Errorf("%s not found", cast.B2S(name))
		}
		return bkt, nil
	}

	return tx.CreateBucketIfNotExists(name)
}

func Bkt2B(b *bbolt.Bucket, name []byte, readonly bool) (*bbolt.Bucket, error) {
	if readonly {
		bkt := b.Bucket(name)
		if bkt == nil {
			return nil, fmt.Errorf("%s not found", cast.B2S(name))
		}
		return bkt, nil
	}
	return b.CreateBucketIfNotExists(name)
}

func Open(path string, mode os.FileMode, options *bbolt.Options) (*bbolt.DB, error) {
	return bbolt.Open(path, mode, options)
}

type Setting struct {
	Names   []string
	Mode    os.FileMode
	Options *bbolt.Options
}

func CacheKV(names ...string) func(setting *Setting) {
	return func(setting *Setting) {
		setting.Names = names
		setting.Options = &bbolt.Options{
			NoGrowSync:   true,
			Timeout:      Default.Timeout,
			FreelistType: Default.FreelistType,
		}
	}

}

func OpenBkt[T any](path string, options ...func(*Setting)) (*Bucket[T], error) {
	var setting Setting
	for _, opt := range options {
		opt(&setting)
	}
	if setting.Mode == 0 {
		setting.Mode = 0644
	}
	if setting.Options == nil {
		setting.Options = Default
	}

	db, err := Open(path, setting.Mode, setting.Options)
	if err != nil {
		return nil, err
	}

	if len(setting.Names) <= 0 {
		setting.Names = []string{"DB_ADMIN"}
	}
	return Pack[T](db, setting.Names...), nil
}
