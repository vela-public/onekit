package bucket

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/mime"
	"strconv"
	"time"
)

type Element struct {
	size  uint64
	ttl   uint64
	mime  string
	chunk []byte
}

func (elem *Element) set(name string, chunk []byte, expire int) {
	var ttl uint64

	if expire > 0 {
		ttl = uint64(time.Now().UnixMilli()) + uint64(expire)
	}

	//如果ttl 为空 第二次传值有时间
	if elem.ttl == 0 {
		elem.ttl = ttl
	}

	elem.mime = name
	elem.size = uint64(len(name))
	elem.chunk = chunk

}

func iEncode(it *Element, v interface{}, expire int) error {
	chunk, name, err := mime.Encode(v)
	if err != nil {
		return err
	}
	it.set(name, chunk, expire)
	return nil
}

func iDecode(it *Element, data []byte) error {
	n := len(data)
	if n == 0 {
		it.mime = mime.NIL
		it.size = 3
		it.chunk = nil
		return nil
	}

	if n < 16 {
		return fmt.Errorf("inavlid item , too small")
	}

	size := binary.BigEndian.Uint64(data[0:8])
	ttl := binary.BigEndian.Uint64(data[8:16])
	now := time.Now().UnixMilli()

	if ttl == 0 || int64(ttl) > now {
		if size+16 == uint64(n) {
			return fmt.Errorf("inavlid item , too big")
		}

		name := data[16 : 16+size]
		chunk := data[size+16:]

		it.size = size
		it.ttl = ttl
		it.mime = cast.B2S(name)
		it.chunk = chunk
		return nil
	}

	it.size = 3
	it.mime = mime.NIL
	it.chunk = it.chunk[:0]
	it.ttl = 0
	return nil
}

func (elem Element) Byte() []byte {
	var buf bytes.Buffer
	size := make([]byte, 8)
	binary.BigEndian.PutUint64(size, elem.size)
	buf.Write(size)

	ttl := make([]byte, 8)
	binary.BigEndian.PutUint64(ttl, elem.ttl)
	buf.Write(ttl)

	buf.WriteString(elem.mime)
	buf.Write(elem.chunk)
	return buf.Bytes()
}

func (elem Element) Decode() (interface{}, error) {
	if elem.mime == "" {
		return nil, fmt.Errorf("not found mime type")
	}

	if elem.mime == mime.NIL {
		return nil, nil
	}

	return mime.Decode[any](elem.mime, elem.chunk)
}

func (elem Element) IsNil() bool {
	return elem.size == 0 || elem.mime == mime.NIL
}

func (elem *Element) incr(v float64, expire int) (sum float64) {
	num, err := elem.Decode()
	if err != nil {
		goto NEW
	}

	switch n := num.(type) {
	case nil:
		sum = v
	case float64:
		sum = n + v
	case float32:
		sum = float64(n) + v
	case int:
		sum = float64(n) + v
	case int8:
		sum = float64(n) + v
	case int16:
		sum = float64(n) + v
	case int32:
		sum = float64(n) + v
	case int64:
		sum = float64(n) + v
	case uint:
		sum = float64(n) + v
	case uint8:
		sum = float64(n) + v
	case uint16:
		sum = float64(n) + v
	case uint32:
		sum = float64(n) + v
	case uint64:
		sum = float64(n) + v
	case string:
		nf, _ := strconv.ParseFloat(n, 10)
		sum = nf + v
	case []byte:
		nf, _ := strconv.ParseFloat(cast.B2S(n), 10)
		sum = nf + v

	default:
		sum = v
	}

NEW:
	chunk, name, _ := mime.Encode(sum)
	elem.set(name, chunk, expire)
	return
}
