package libkit

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileExists checks if the file exists in the provided path
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// OpenOrCreate opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDWR.
// If there is an error, it'll create the named file with mode 0666.
// If successful, methods on the returned File can be used for I/O;
// the associated file descriptor has mode O_RDWR.
// If there is an error, it will be of type *PathError.
// Note: The file gets created only if the target directory exists
func OpenOrCreateFile(name string) (*os.File, error) {
	if FileExists(name) {
		return os.OpenFile(name, os.O_RDWR, 0666)
	}
	return os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
}

// CreateIfNotExists creates a file or a directory only if it does not already exist.
func CreateIfNotExists(path string, isDir bool) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if isDir {
			return os.MkdirAll(path, 0755)
		}

		if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
			return e
		}

		f, e := os.OpenFile(path, os.O_CREATE, 0755)
		if e != nil {
			return err
		}
		_ = f.Close()
		return nil
	}
	return nil
}

func ReadlineFuncText(path string, fn func(string) (stop bool)) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if e := scanner.Err(); e != nil {
			if e == io.EOF {
				return nil
			}
			return e
		}

		if fn(scanner.Text()) {
			return nil
		}
	}

	return nil
}

func ReadlineFunc(path string, fn func(string) (stop bool, err error)) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if e := scanner.Err(); e != nil {
			if e == io.EOF {
				return nil
			}
			return e
		}

		text := scanner.Text()
		text = strings.TrimSpace(text)
		if len(text) == 0 {
			continue
		}
		if text[0] == '#' {
			continue
		}

		if s, e := fn(text); s {
			return e
		}
	}

	return nil
}
