package wasm

import (
	"unsafe"
)

type Payload struct {
	Code  uint32
	Size  uint32
	Large uint64
	Data  []byte
}

/*


 */

func (p *Payload) Text() []byte {
	buf := make([]byte, 16)
	ptr := unsafe.Pointer(&p.Code)
	for k := uint32(0); k < 4; k++ {
		buf[k] = *(*byte)(unsafe.Add(ptr, k))
	}

	ptr = unsafe.Pointer(&p.Size)
	for k := uint32(0); k < 4; k++ {
		buf[k+4] = *(*byte)(unsafe.Add(ptr, k))
	}

	ptr = unsafe.Pointer(&p.Large)
	for k := uint32(0); k < 8; k++ {
		buf[k+8] = *(*byte)(unsafe.Add(ptr, k))
	}

	text := make([]byte, 0, len(p.Data)+16)
	text = append(text, buf...)
	text = append(text, p.Data...)
	return text
}

func (p *Payload) Build(text []byte) {
	n := len(text)
	if n < 16 {
		return
	}

	ct := text[:4]
	code := (*uint32)(unsafe.Pointer(&ct[0]))

	sz := text[4:8]
	size := (*uint32)(unsafe.Pointer(&sz[0]))

	lz := text[8:16]
	large := (*uint64)(unsafe.Pointer(&lz[0]))

	p.Code = *code
	p.Size = *size
	p.Large = *large
	p.Data = text[16:]
}

func NewPayload(code uint32, data []byte) *Payload {
	return &Payload{
		Code: code,
		Size: uint32(len(data)),
		Data: data,
	}
}
