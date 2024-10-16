package libkit

import (
	"bufio"
	"io"
	"os"
	"strings"
)

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
