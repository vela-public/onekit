package bucket

import (
	"fmt"
	"github.com/vela-public/onekit/cast"
	"go.etcd.io/bbolt"
)

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
