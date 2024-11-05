package filekit

const (
	Nothing = iota
	Running
	Paused
	Stopped
	Cleaned
	Done
)

type ErrNo int8
