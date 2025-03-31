package filekit

import (
	"bufio"
	"bytes"
	"fmt"
	"sync/atomic"
)

type LineFSM struct {
	tail    *FileTail
	scanner *bufio.Reader
	next    bool
	err     error
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

func (fsm *LineFSM) Read() (string, error) {
	fsm.tail.Wait()
	var buff bytes.Buffer
	defer fsm.Reset()
	var size int32

repeat:
	data, next, err := fsm.scanner.ReadLine()
	fsm.err = err
	fsm.next = next

	if sz := len(data); sz > 0 {
		buff.Write(data)
		if atomic.AddInt32(&size, int32(sz)) > 1024*1024 {
			return buff.String(), fmt.Errorf("line size too large %d", size)
		}
	}

	if err == nil && next {
		goto repeat
	}

	return buff.String(), err

}
