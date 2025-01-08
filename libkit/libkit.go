package libkit

import (
	"bufio"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Merge[T comparable](a []T, b T) []T {
	sz := len(a)
	if sz == 0 {
		return []T{b}
	}

	for i := 0; i < sz; i++ {
		v := a[i]
		if v == b {
			return a
		}
	}
	return append(a, b)
}

func Merges[T comparable](a []T, b ...T) []T {
	sz := len(a)
	if sz == 0 {
		return b
	}
	for _, v := range b {
		a = Merge(a, v)
	}
	return a
}

// Dedupe removes duplicates from a slice of elements preserving the order
func Dedupe[T comparable](inputSlice []T) (result []T) {
	seen := make(map[T]struct{})
	for _, inputValue := range inputSlice {
		if _, ok := seen[inputValue]; !ok {
			seen[inputValue] = struct{}{}
			result = append(result, inputValue)
		}
	}

	return
}

// PickRandom item from a slice of elements
func PickRandom[T any](v []T) T {
	return v[rand.Intn(len(v))]
}

// Contains if a slice contains an element
func Contains[T comparable](inputSlice []T, element T) bool {
	for _, inputValue := range inputSlice {
		if inputValue == element {
			return true
		}
	}

	return false
}

// Contains if a slice contains an element
func In[T comparable](inputSlice []T, element T) bool {
	for _, inputValue := range inputSlice {
		if inputValue == element {
			return true
		}
	}

	return false
}

// Diff calculates the extra elements between two sequences
func Diff[V comparable](s1, s2 []V) (extraS1, extraS2 []V) {
	s1Len := len(s1)
	s2Len := len(s2)

	visited := make([]bool, s2Len)
	for i := 0; i < s1Len; i++ {
		element := s1[i]
		found := false
		for j := 0; j < s2Len; j++ {
			if visited[j] {
				continue
			}
			if s2[j] == element {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			extraS1 = append(extraS1, element)
		}
	}

	for j := 0; j < s2Len; j++ {
		if visited[j] {
			continue
		}
		extraS2 = append(extraS2, s2[j])
	}

	return
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

func Wait() os.Signal {
	chn := make(chan os.Signal, 1)
	signal.Notify(chn, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	s := <-chn
	return s
}
