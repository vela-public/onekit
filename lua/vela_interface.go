package lua

import "io"

type LVFace interface {
	ToLValue() LValue
}

type Writer interface {
	io.Writer
}

type IO interface {
	io.Writer
	io.Reader
}

type Reader interface {
	io.Reader
}

type Closer interface {
	io.Closer
}

type ReaderCloser interface {
	io.Reader
	io.Closer
}

type WriterCloser interface {
	io.Writer
	io.Closer
}
type Preloader interface {
	Set(string, LValue)
	SetGlobal(string, LValue)
	Get(string) LValue
	Global(string) LValue
}
