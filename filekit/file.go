package filekit

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
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

func Md5(path string) (string, error) {
	fd, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %v error %v", path, err)
	}
	defer fd.Close()

	hub := md5.New()
	io.Copy(hub, fd)
	return hex.EncodeToString(hub.Sum(nil)), nil
}

// CopyFile copies from src to dst until either EOF is reached
// on src or an error occurs. It verifies src exists and removes
// the dst if it exists.
func CopyFile(src, dst string) (int64, error) {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)
	if cleanSrc == cleanDst {
		return 0, nil
	}
	sf, err := os.Open(cleanSrc)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	if err := os.Remove(cleanDst); err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	df, err := os.Create(cleanDst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

// Copy copies a src file to a dst file where src and dst are regular files.
func Copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// ReadSymlinkedDirectory returns the target directory of a symlink.
// The target of the symbolic link may not be a file.
func ReadSymlinkedDirectory(path string) (string, error) {
	var realPath string
	var err error
	if realPath, err = filepath.Abs(path); err != nil {
		return "", fmt.Errorf("unable to get absolute path for %s: %s", path, err)
	}
	if realPath, err = filepath.EvalSymlinks(realPath); err != nil {
		return "", fmt.Errorf("failed to canonicalise path for %s: %s", path, err)
	}
	realPathInfo, err := os.Stat(realPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat target '%s' of '%s': %s", realPath, path, err)
	}
	if !realPathInfo.Mode().IsDir() {
		return "", fmt.Errorf("canonical path points to a file '%s'", realPath)
	}
	return realPath, nil
}

// CreateIfNotExists creates a file or a directory only if it does not already exist.
func CreateIfNotExists(path string, isDir bool) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if isDir {
				return os.MkdirAll(path, 0755)
			}
			if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
				return err
			}
			if f, e := os.OpenFile(path, os.O_CREATE, 0755); e != nil {
				return e
			} else {
				_ = f.Close()
			}
			return nil
		}
	}
	return nil
}
