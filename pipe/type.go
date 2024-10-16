package pipe

const (
	Single HandleType = iota + 1
	ReuseCo
)

type HandleType uint8

type Invoker interface {
	Invoke(v interface{}) error
}

type Bytes interface {
	Bytes() []byte
}
