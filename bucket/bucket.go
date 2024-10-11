package bucket

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/mime"
	"go.etcd.io/bbolt"
	"os"
)

var Default = bbolt.DefaultOptions

type Tx = bbolt.Tx

type Bucket[T any] struct {
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
