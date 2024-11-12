package errkit

import "fmt"

var ErrNotSupported = fmt.Errorf("not supported")

const (
	Undefined ErrNo = 1 << iota
	Succeed
	Failed
	NotFound
	Conflict
	Invalid
	Unauthorized
	Forbidden
	Closed
	Timeout
	Temporary
	building
)

type ErrNo int

func (en ErrNo) Error() string {
	switch en {
	case Succeed:
		return "succeed"
	case Closed:
		return "closed"
	case Timeout:
		return "timeout"
	case Temporary:
		return "temporary"
	case Forbidden:
		return "forbidden"
	case Conflict:
		return "conflict"
	case Failed:
		return "failed"
	case NotFound:
		return "not found"
	case Unauthorized:
		return "unauthorized"
	case Invalid:
		return "invalid"
	case Undefined:
		return "undefined"
	case building:
		return "building"
	}

	return "unknown"
}

func (en ErrNo) Code() int {
	return int(en)
}

func (en ErrNo) Have(v ErrNo) bool {
	return en&v > 0
}

func (en ErrNo) Or(v ErrNo) ErrNo {
	return en | v
}

func (en ErrNo) And(v ErrNo) ErrNo {
	return en & v
}

func (en ErrNo) Binary() string {
	return fmt.Sprintf("%032b", en)
}
