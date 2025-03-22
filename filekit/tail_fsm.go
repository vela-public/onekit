package filekit

import (
	"bufio"
	"bytes"
)

type LineFSM struct {
	tail  *FileTail
	buff  *bufio.Reader
	parts [][]byte
	next  bool
	err   error
}

func (fsm *LineFSM) Next() bool {
	return fsm.next
}
func (fsm *LineFSM) UnwrapErr() error {
	return fsm.err
}

func (fsm *LineFSM) Reset() {
	if sz := len(fsm.parts); sz > 0 {
		return
	}
	fsm.parts = fsm.parts[:0]
	fsm.err = nil
	fsm.next = false
}

func (fsm *LineFSM) Text() []byte {
	text := bytes.Join(fsm.parts, nil)
	fsm.Reset()
	return text
}

func (fsm *LineFSM) Read() {
	fsm.tail.Wait()
	data, next, err := fsm.buff.ReadLine()
	fsm.err = err
	fsm.next = next

	if len(data) > 0 {
		dst := make([]byte, len(data))
		copy(dst, data)
		fsm.parts = append(fsm.parts, dst)
	}
}
