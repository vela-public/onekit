package bucket

const (
	OK              = 2
	built           = 1
	Init            = 0
	NotFound        = -1
	TooSmall        = -2
	TooBig          = -3
	Expired         = -4
	MimeDecodeError = -5
	MimeEncodeError = -6
	TypeError       = -7
	InternalError   = -8
)

type ErrNo int
