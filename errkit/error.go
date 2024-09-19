package errkit

import (
	"bytes"
)

type errKV struct {
	key string
	err error
}

type JoinError struct {
	data []errKV
}

func (e *JoinError) Len() int {
	return len(e.data)
}

func (e *JoinError) Error() string {
	n := e.Len()
	if n == 0 {
		return ""
	}

	var buff bytes.Buffer
	for i := 0; i < n; i++ {
		if i != 0 {
			buff.WriteByte('\n')
		}

		item := e.data[i]
		if item.key != "" {
			buff.WriteString(item.key)
			buff.WriteByte(':')
		}
		buff.WriteString(item.err.Error())

	}

	return buff.String()
}

func (e *JoinError) Try(key string, err error) {
	if err == nil {
		return
	}
	e.data = append(e.data, errKV{key, err})
}

func (e *JoinError) Wrap() error {
	if len(e.data) == 0 {
		return nil
	}

	return e
}

func New() *JoinError {
	return &JoinError{}
}
