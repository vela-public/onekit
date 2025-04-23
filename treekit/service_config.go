package treekit

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/vela-public/onekit/cast"
	"hash/crc32"
	"io"
	"os"
	"strings"
)

type MicoServiceConfig struct {
	// task ID
	ID int64

	// task key
	Key string

	// task code hash
	Hash string

	// source
	Source []byte

	// source size
	size int64

	// task disable
	Disable bool

	// Is it loaded locally
	Dialect bool

	//filepath
	Path string

	//modify time
	MTime int64
}

func (msc *MicoServiceConfig) NewReader() io.Reader {
	return bytes.NewReader(msc.Source)
}

func (msc *MicoServiceConfig) verify() error {
	if e := Name(msc.Key); e != nil {
		return e
	}

	if msc.ID == 0 {
		return fmt.Errorf("task.id must be greater than 0")
	}

	if len(msc.Source) == 0 {
		return fmt.Errorf("empty document")
	}

	return nil

}

func NewFile(key string, path string) (*MicoServiceConfig, error) {
	cfg := &MicoServiceConfig{
		ID:      int64(crc32.ChecksumIEEE(cast.S2B(key))),
		Key:     key,
		Path:    path,
		Dialect: true,
	}

	fd, err := os.Open(path)
	if err != nil {
		return cfg, err
	}

	st, err := fd.Stat()
	if err == nil {
		cfg.MTime = st.ModTime().Unix()
	}

	m5 := md5.New()
	buf := bytes.NewBuffer(nil)

	w := io.MultiWriter(m5, buf)
	size, err := io.Copy(w, fd)
	if err != nil && err != io.EOF {
		return cfg, err
	}
	cfg.size = size
	cfg.Source = buf.Bytes()
	cfg.Hash = fmt.Sprintf("%x", m5.Sum(nil))

	return cfg, cfg.verify()
}

func NewText(key string, data string, mtime int64) (*MicoServiceConfig, error) {
	cfg := &MicoServiceConfig{
		ID:      int64(crc32.ChecksumIEEE(cast.S2B(key))),
		Key:     key,
		Path:    "memory",
		Dialect: true,
		MTime:   mtime,
	}

	fd := strings.NewReader(data)
	m5 := md5.New()
	buf := bytes.NewBuffer(nil)

	w := io.MultiWriter(m5, buf)
	size, err := io.Copy(w, fd)
	if err != nil && err != io.EOF {
		return cfg, err
	}
	cfg.size = size
	cfg.Source = buf.Bytes()
	cfg.Hash = fmt.Sprintf("%x", m5.Sum(nil))

	return cfg, cfg.verify()
}

func NewConfig(key string, options ...func(config *MicoServiceConfig)) (*MicoServiceConfig, error) {
	cfg := &MicoServiceConfig{
		Key:     key,
		Dialect: false,
	}

	for _, option := range options {
		option(cfg)
	}

	if cfg.ID == 0 {
		cfg.ID = int64(crc32.ChecksumIEEE(cast.S2B(key)))
	}

	if cfg.Hash == "" {
		m5 := md5.New()
		m5.Write(cfg.Source)
		cfg.Hash = fmt.Sprintf("%x", m5.Sum(nil))
	}
	return cfg, cfg.verify()

}
