package bucket

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/mime"
	"time"
)

type Element[T any] struct {
	now  int64
	flag ErrNo
	data T
	info error

	size uint64 //mime name size
	ttl  uint64
	mime string
	text []byte
}

func (elem *Element[T]) ErrNo() ErrNo {
	return elem.flag
}

func (elem *Element[T]) Mime() string {
	return elem.mime
}

func (elem *Element[T]) len() int {
	return 16 + len(elem.mime) + len(elem.text)
}
func (elem *Element[T]) Now() int64 {
	if elem.now == 0 {
		elem.now = time.Now().UnixMilli()
	}
	return elem.now
}

func (elem *Element[T]) fill(name string, text []byte, expire int) {
	var ttl uint64

	if expire > 0 {
		ttl = uint64(elem.Now()) + uint64(expire)
	}

	//如果ttl 为空 第二次传值有时间
	if elem.ttl == 0 {
		elem.ttl = ttl
	}

	elem.mime = name
	elem.size = uint64(len(name))
	elem.text = text
}

func (elem *Element[T]) Set(t T, expire int) {
	chunk, name, err := mime.Encode(t)
	if err != nil {
		elem.flag = MimeEncodeError
		elem.info = err
		return
	}
	elem.fill(name, chunk, expire)
	elem.data = t
	elem.flag = OK
}

func (elem *Element[T]) Upsert(t T, expire int) error {
	chunk, name, err := mime.Encode(t)
	if err != nil {
		return err
	}

	elem.mime = name
	elem.size = uint64(len(name))
	elem.text = chunk

	if elem.ttl == 0 && expire == 0 {
		return nil
	}

	if elem.ttl == 0 && expire > 0 { // 设置过期
		elem.ttl = uint64(elem.Now()) + uint64(expire)
		return nil
	}

	if elem.Now() >= int64(elem.ttl) { // 已经过期
		elem.ttl = uint64(elem.Now()) + uint64(expire)
		return nil
	}

	if elem.ttl-uint64(elem.Now()) > uint64(expire) { // 刷新过期时间
		elem.ttl = uint64(elem.Now()) + uint64(expire)
		return nil
	}

	return nil
}

func (elem *Element[T]) Expired() bool {
	if elem.ttl == 0 || int64(elem.ttl) > elem.now {
		return false
	}
	return true
}

func (elem *Element[T]) Build(data []byte) {
	sz := len(data)
	if sz == 0 {
		elem.flag = NotFound
		elem.info = fmt.Errorf("not found")
		return
	}

	if sz < 16 {
		elem.flag = TooSmall
		elem.info = fmt.Errorf("bad element , too small")
		return
	}

	n := binary.BigEndian.Uint64(data[:8])

	if n+16 == uint64(sz) {
		elem.flag = TooBig
		elem.info = fmt.Errorf("bad element , too big")
		return
	}

	name := cast.B2S(data[16 : 16+n])
	text := data[n+16:]
	ttl := binary.BigEndian.Uint64(data[8:16])
	now := time.Now().UnixMilli()

	if ttl == 0 || int64(ttl) > now {
		elem.size = n
		elem.ttl = ttl
		elem.mime = name
		elem.text = text
		elem.now = now
		return
	}

	elem.flag = Expired
	elem.mime = name
	elem.size = n
	elem.text = nil
	elem.ttl = ttl
	elem.now = now
}

func (elem *Element[T]) Text() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, elem.len()))
	size := make([]byte, 8)
	binary.BigEndian.PutUint64(size, elem.size)
	buf.Write(size)

	ttl := make([]byte, 8)
	binary.BigEndian.PutUint64(ttl, elem.ttl)
	buf.Write(ttl)

	buf.WriteString(elem.mime)
	buf.Write(elem.text)
	return buf.Bytes()
}

func (elem *Element[T]) Value() T {
	if elem.flag != 0 {
		return elem.data
	}
	t, _ := elem.Unwrap()
	return t
}

func (elem *Element[T]) UnwrapErr() error {
	return elem.info
}

func (elem *Element[T]) Unwrap() (t T, e error) {
	if elem.flag != 0 {
		return elem.data, elem.info
	}

	var v any
	de, ok := any(elem.data).(mime.Decoder)
	if ok {
		v, e = de.MimeDecode(elem.text)
	} else {
		v, e = mime.Decode(elem.mime, elem.text)
	}

	if e != nil {
		elem.flag = MimeDecodeError
		elem.info = e
		return
	}

	switch vt := v.(type) {
	case T:
		elem.flag = OK
		elem.data = vt
		return vt, nil
	case *T:
		elem.flag = OK
		elem.data = *vt
		return *vt, nil
	}

	elem.flag = TypeError
	elem.info = fmt.Errorf("bad element , type mismatch must:%T got:%T", elem.data, v)
	return elem.data, elem.info
}
