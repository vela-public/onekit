package filekit

import (
	"bufio"
	"bytes"
)

type LineFSM struct {
	tail   *FileTail
	reader *bufio.Reader
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
	fsm.err = nil
	fsm.next = false
}

func (fsm *LineFSM) Read() ([]byte, error) {
	fsm.tail.Wait()
	var buff bytes.Buffer
	defer fsm.Reset()

repeat:
	data, next, err := fsm.reader.ReadLine()
	fsm.err = err
	fsm.next = next

	if sz := len(data); sz > 0 {
		buff.Write(data)
	}

	if err == nil && next {
		goto repeat
	}

	return buff.Bytes(), err

}
