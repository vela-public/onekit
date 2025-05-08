package render

import (
	"github.com/vela-public/onekit/cast"
	"io"
	"os"
)

type FileTemplate struct {
	Path  string `lua:"path"`
	Auto  bool   `lua:"auto"`
	Max   int    `lua:"max"`
	Cache bool   `lua:"cache"`
	MTime int64  `lua:"-"`
	Text  string `lua:"-"`
}

func (ft *FileTemplate) ReadFile() ([]byte, error) {
	file, err := os.Open(ft.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if ft.Max <= 0 {
		return io.ReadAll(file)
	}

	stat, err := file.Stat()
	if err == nil {
		ft.MTime = stat.ModTime().Unix()
	}

	buf := make([]byte, ft.Max)

	n, err := file.Read(buf)
	return buf[:n], err
}

func (ft *FileTemplate) Reader() (string, bool) {
	if !ft.Cache {
		buf, err := ft.ReadFile()
		if err != nil {
			return err.Error(), true
		}
		return cast.B2S(buf), true
	}

	if ft.Text == "" {
		buf, err := ft.ReadFile()
		if err != nil {
			return err.Error(), true
		}
		ft.Text = cast.B2S(buf)
		return ft.Text, true
	}

	stat, err := os.Stat(ft.Path)
	if err != nil {
		return err.Error(), true
	}

	mtime := stat.ModTime().Unix()
	if mtime != ft.MTime {
		buf, e := ft.ReadFile()
		if e != nil {
			return e.Error(), true
		}
		ft.Text = cast.B2S(buf)
		return ft.Text, true
	}

	return ft.Text, false
}
