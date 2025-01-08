package taskit

import "bytes"

const (
	defined ErrNo = 1 << iota
	Succeed
	Stopped
	Failed
	Reset
	Reload
)

var ErrNoMap = map[ErrNo]string{
	defined: "defined",
	Succeed: "succeed",
	Stopped: "stopped",
	Failed:  "failed",
	Reset:   "reset",
	Reload:  "reload",
}

type ErrNo uint8

func (en ErrNo) String() string { return ErrNoMap[en] }
func (en ErrNo) Text() string {
	buf := bytes.NewBuffer(make([]byte, 0, 50))
	offset := 0
	concat := func(s string) {
		if offset > 0 {
			buf.WriteString("|")
		}
		buf.WriteString(s)
		offset++
	}

	switch {
	case en == 0:
		concat("zero")
	case en&defined == defined:
		concat("defined")
	case en&Succeed == Succeed:
		concat("succeed")
	case en&Stopped == Stopped:
		concat("stopped")
	case en&Failed == Failed:
		concat("failed")
	case en&Reset == Reset:
		concat("reset")
	default:
		concat("unknown")
	}

	return buf.String()
}
