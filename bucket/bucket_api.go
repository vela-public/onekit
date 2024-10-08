package bucket

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/mime"
	"go.etcd.io/bbolt"
	"time"
)

const (
	STOP = iota + 1
	CONTINUE
	REMOVE
	ERROR
)

type ForEachFSM uint8

func (b *Bucket) ForEach(callback func(key, val []byte) (error, ForEachFSM)) error {
	err := b.db.Batch(func(tx *bbolt.Tx) error {
		bt, err := b.unpack(tx, false)
		if err != nil {
			return err
		}
		var deletes [][]byte
		err = bt.ForEach(func(k, v []byte) error {
			elem := &Element{}
			if e := iDecode(elem, v); e != nil {
				deletes = append(deletes, k)
				return nil
			}

			e, fsm := callback(k, elem.chunk)
			switch fsm {
			case STOP:
				return e
			case CONTINUE:
				return nil
			case REMOVE:
				deletes = append(deletes, k)
				return nil
			default:
				return nil
			}
		})

		for _, k := range deletes {
			err = bt.Delete(k)
		}

		if err != nil {
			return err
		}
		return nil
	})

	return err
}

func (b *Bucket) Store(key string, v interface{}, expire int) error {
	err := b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return err
		}
		var it Element
		err = iEncode(&it, v, expire)
		if err != nil {
			return err
		}
		bbt.Put(cast.S2B(key), it.Byte())
		return nil
	})
	return err
}

func (b *Bucket) Load(key string) (Element, error) {
	it := Element{}

	err := b.db.View(func(tx *Tx) error {
		bbt, err := b.unpack(tx, true)
		if err != nil {
			return err
		}

		data := bbt.Get(cast.S2B(key))
		err = iDecode(&it, data)
		return err
	})

	if err != nil {
		return it, err
	}

	return it, nil
}

func (b *Bucket) Replace(key string, v interface{}, expire int) error {
	return b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return err
		}
		kb := cast.S2B(key)
		data := bbt.Get(kb)

		chunk, name, err := mime.Encode(v)
		if err != nil {
			return err
		}

		var it Element
		err = iDecode(&it, data)
		if err != nil {
			it.set(name, chunk, expire)
			return err
		}

		it.mime = name
		it.chunk = chunk
		return bbt.Put(kb, it.Byte())
	})
}

func (b *Bucket) Delete(key string) error {
	return b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return nil
		}
		return bbt.Delete(cast.S2B(key))
	})
}

func (b *Bucket) DeleteBucket(nb string) error {
	return b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return nil
		}
		return bbt.DeleteBucket(cast.S2B(nb))
	})
}

func (b *Bucket) Get(key string) (interface{}, error) {
	it, err := b.Load(key)
	if err != nil {
		return nil, err
	}
	return it.Decode()
}

func (b *Bucket) Int(key string) int {
	val, err := b.Get(key)
	if err != nil {
		return 0
	}
	return cast.ToInt(val)
}

func (b *Bucket) Int64(key string) int64 {
	val, err := b.Get(key)
	if err != nil {
		return 0
	}
	return cast.ToInt64(val)
}

func (b *Bucket) Bool(key string) bool {
	val, err := b.Get(key)
	if err != nil {
		return false
	}

	return cast.ToBool(val)
}

func (b *Bucket) unpack(tx *Tx, readonly bool) (*bbolt.Bucket, error) {
	var bbt *bbolt.Bucket
	var err error

	if b.db == nil {
		return nil, errors.New("not found database")
	}

	n := len(b.chains)
	if n < 1 {
		return nil, errors.New("not found bucket")
	}

	bbt, err = Tx2B(tx, b.chains[0], readonly)
	if n == 1 {
		return bbt, err
	}

	//如果报错
	if err != nil {
		return bbt, err
	}

	for i := 1; i < n; i++ {
		bbt, err = Bkt2B(bbt, b.chains[i], readonly)
		if err != nil {
			return nil, err
		}
	}

	return bbt, nil
}

func (b *Bucket) Push(key string, val []byte, expire int64) error {
	return b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return err
		}

		var ttl uint64
		if expire <= 0 {
			ttl = 0
		} else {
			ttl = uint64(time.Now().Unix() + expire)
		}

		it := Element{
			mime:  mime.BYTES,
			size:  uint64(len(mime.BYTES)),
			ttl:   ttl,
			chunk: val,
		}
		return bbt.Put(cast.S2B(key), it.Byte())
	})
}

func (b *Bucket) Value(key string) ([]byte, error) {
	it, err := b.Load(key)
	if err != nil {
		return nil, err
	}

	switch it.mime {
	case mime.NIL:
		return nil, nil
	case mime.STRING, mime.BYTES:
		return it.chunk, nil

	default:
		return nil, fmt.Errorf("%s not bytes , got %s", key, it.mime)
	}
}

func (b *Bucket) Json() *Bucket {
	b.export = "json"
	return b
}

func (b *Bucket) Names() string {
	return string(bytes.Join(b.chains, []byte(",")))
}

func (b *Bucket) String() string {
	return fmt.Sprintf("%p", b)
}
