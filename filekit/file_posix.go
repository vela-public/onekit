//go:build linux || darwin || freebsd || netbsd || openbsd
// +build linux darwin freebsd netbsd openbsd

package filekit

import (
	"os"
)

func OpenFile(filename string) (*os.File, error) {
	return os.Open(filename)
}
