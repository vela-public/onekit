package bucket

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"go.etcd.io/bbolt"
)

const (
	STOP = iota + 1
	CONTINUE
	REMOVE
	ERROR
)

type ForEachFSM uint8

func (b *Bucket[T]) WithText(key string, fn func() string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key is empty")
	}

	var ret []byte
	cp := func(dat []byte) {
		sz := len(dat)
		if sz == 0 {
			return
		}
		ret = make([]byte, sz)
		copy(ret, dat)
	}

	err := b.db.Batch(func(tx *bbolt.Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return err
		}

		k := []byte(key)
		data := bbt.Get(k)
		if len(data) > 0 {
			cp(data)
			return nil
		}

		ret = cast.S2B(fn())
		if len(ret) == 0 {
			return nil
		}
		return bbt.Put(k, ret)
	})

	return cast.B2S(ret), err
}

func (b *Bucket[T]) SetText(key string, text string) error {
	if key == "" || len(text) == 0 {
		return nil
	}

	err := b.db.Update(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return err
		}

		return bbt.Put(cast.S2B(key), cast.S2B(text))
	})
	return err
}

func (b *Bucket[T]) GetText(key string, e ...func(error)) string {
	if key == "" {
		return ""
	}

	err := b.CreateBucketIfNotExists()
	bad := func() {
		if sz := len(e); sz > 0 {
			for i := 0; i < sz; i++ {
				e[i](err)
			}
		}
	}

	if err != nil {
		bad()
		return ""
	}

	var data []byte
	err = b.db.View(func(tx *Tx) error {
		bbt, ei := b.unpack(tx, true)
		if ei != nil {
			return ei
		}

		dat := bbt.Get(cast.S2B(key))
		if sz := len(dat); sz > 0 {
			data = make([]byte, sz)
			copy(data, dat)
		}
		return nil
	})

	if err != nil {
		bad()
		return ""
	}

	return cast.B2S(data)
}

func (b *Bucket[T]) CreateBucketIfNotExists() error {
	var err error
	b.once.Do(func() {
		err = b.db.Update(func(tx *bbolt.Tx) error {
			_, e := b.unpack(tx, false)
			if e != nil {
				return e
			}
			return nil
		})
	})
	return err

}

func (b *Bucket[T]) unpack(tx *Tx, readonly bool) (*bbolt.Bucket, error) {
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

func (b *Bucket[T]) ForEach(callback func(key, val []byte) (error, ForEachFSM)) error {
	err := b.db.Batch(func(tx *bbolt.Tx) error {
		bt, err := b.unpack(tx, false)
		if err != nil {
			return err
		}
		err = bt.ForEach(func(k, v []byte) error {
			elem := new(Element[T])
			elem.Build(v)
			switch elem.flag {
			case TooSmall, TooBig, Expired, NotFound:
				return bt.Delete(k)
			case Built:
				e, fsm := callback(k, elem.text)
				switch fsm {
				case STOP:
					return e
				case CONTINUE:
					return nil
				case REMOVE:
					_ = bt.Delete(k)
					return nil
				default:
					return nil
				}
			}

			return nil
		})
		return err
	})
	return err
}

func (b *Bucket[T]) Set(key string, v T, expire int) error {
	err := b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return err
		}
		elem := new(Element[T])
		elem.Set(v, expire)
		if elem.flag != OK {
			return elem.info
		}
		return bbt.Put(cast.S2B(key), elem.Text())
	})
	return err
}

func (b *Bucket[T]) Get(key string) *Element[T] {
	elem := new(Element[T])
	var data []byte
	err := b.db.View(func(tx *Tx) error {
		bbt, err := b.unpack(tx, true)
		if err != nil {
			return err
		}

		dat := bbt.Get(cast.S2B(key))
		if sz := len(dat); sz > 0 {
			data = make([]byte, sz)
			copy(data, dat)
		}
		return nil
	})

	if err != nil {
		elem.flag = -1
		elem.info = err
		return elem
	}

	elem.Build(data)
	return elem
}

func (b *Bucket[T]) Upsert(key string, v T, expire int) *Element[T] {
	elem := new(Element[T])
	err := b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			elem.flag = InternalError
			elem.info = err
			return err
		}

		kb := cast.S2B(key)
		data := bbt.Get(kb)
		elem.Build(data)
		switch elem.flag {
		case NotFound, TooBig, TooSmall:
			elem.Set(v, expire)
		case Built:
			if e := elem.Upsert(v, expire); e != nil {
				return e
			}
		}
		return bbt.Put(kb, elem.Text())
	})

	if err == nil {
		elem.flag = InternalError
		elem.info = err
	} else {
		elem.flag = OK
		elem.data = v
	}

	return elem
}

func (b *Bucket[T]) Delete(key string) error {
	err := b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return nil
		}

		return bbt.Delete(cast.S2B(key))
	})

	return err
}

func (b *Bucket[T]) DeleteBucket(nb string) error {
	return b.db.Batch(func(tx *Tx) error {
		bbt, err := b.unpack(tx, false)
		if err != nil {
			return nil
		}
		return bbt.DeleteBucket(cast.S2B(nb))
	})
}

func (b *Bucket[T]) Path() string {
	return string(bytes.Join(b.chains, []byte(",")))
}
