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

func (en ErrNo) String() string { return en.Text() }
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

	if en == 0 {
		concat("zero")
	}

	if en&defined == defined {
		concat("defined")
	}
	if en&Succeed == Succeed {
		concat("succeed")
	}
	if en&Stopped == Stopped {
		concat("stopped")
	}
	if en&Failed == Failed {
		concat("failed")
	}
	if en&Reset == Reset {
		concat("reset")
	}

	return buf.String()
}
