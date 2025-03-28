package filekit

import (
	"bufio"
	"bytes"
)

type LineFSM struct {
	tail   *FileTail
	reader *bufio.Reader
	buffer bytes.Buffer
	next   bool
	err    error
}

func (fsm *LineFSM) Next() bool {
	return fsm.next
}
func (fsm *LineFSM) UnwrapErr() error {
	return fsm.err
}

func (fsm *LineFSM) Reset() {
	fsm.buffer.Reset()
	fsm.err = nil
	fsm.next = false
}

func (fsm *LineFSM) Text() []byte {
	return fsm.buffer.Bytes()
}

func (fsm *LineFSM) Read() {
	fsm.tail.Wait()
	data, next, err := fsm.reader.ReadLine()
	fsm.err = err
	fsm.next = next

	if len(data) > 0 {
		fsm.buffer.Write(data)
	}
}
