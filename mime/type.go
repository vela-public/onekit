package mime

type Encoder interface {
	MimeEncode() ([]byte, error)
}

type Decoder interface {
	MimeDecoder([]byte) error
}
