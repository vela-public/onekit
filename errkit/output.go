package errkit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Output(filename string) (func(string, ...interface{}), *os.File) {
	exe, _ := os.Executable()
	path := filepath.Dir(exe) + filename

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("errkit output open file %v\n", err)
	}

	return func(format string, args ...interface{}) {
		header := fmt.Sprintf("%s %s", time.Now().Format("2006-01-02 15:04:05"), format)
		if file == nil {
			fmt.Printf(header, args...)
			return
		}
		file.WriteString(fmt.Sprintf(header, args...))
	}, file

}
