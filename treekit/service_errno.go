package treekit

const (
	Undefined SerErrNo = 1 << iota
	Register
	Waking
	Running
	Panic
	Fail
	Update
	Disable
	Empty
)

var SerErrNoMap = map[SerErrNo]string{
	Undefined: "undefined",
	Register:  "register",
	Waking:    "waking",
	Running:   "running",
	Panic:     "panic",
	Fail:      "fail",
	Update:    "update",
	Disable:   "disable",
	Empty:     "empty",
}

func (s SerErrNo) String() string {
	return SerErrNoMap[s]
}

type SerErrNo uint32
