package bucket

import (
	"github.com/vela-public/onekit/cast"
	"go.etcd.io/bbolt"
)

type Tx = bbolt.Tx

type Bucket struct {
	db     *bbolt.DB
	chains [][]byte
	export string
}

func Pack(db *bbolt.DB, names ...string) *Bucket {
	var chains [][]byte

	for _, name := range names {
		chains = append(chains, cast.S2B(name))
	}

	return &Bucket{
		db:     db,
		chains: chains,
	}
}
