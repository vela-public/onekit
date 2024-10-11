package mime

import (
	"fmt"
)

var (
	NotFound = fmt.Errorf("not found mime type")
)

type Decoder interface {
	MimeDecode([]byte) (any, error)
}

type Encoder interface {
	MimeEncode(any) ([]byte, error)
}

type TypeOf interface {
	TypeFor() any
	Decoder
	Encoder
}
