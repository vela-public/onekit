package filekit

import (
	"github.com/vela-public/onekit/bucket"
	"go.etcd.io/bbolt"
)

type Seeker interface {
	Save(file string, offset int64) error
	Find(file string) (int64, error)
}

type SeekDB struct {
	db     *bbolt.DB
	bucket []string
}

func (s *SeekDB) Save(file string, offset int64) error {
	db := bucket.Pack[int64](s.db, s.bucket...)
	return db.Set(file, offset, 0)
}

func (s *SeekDB) Find(file string) (int64, error) {
	db := bucket.Pack[int64](s.db, s.bucket...)
	seek, err := db.Get(file).Unwrap()
	if err != nil {
		return 0, err
	}
	return seek, nil
}

type SeekMem struct {
	seek map[string]int64
}

func (s *SeekMem) Save(file string, offset int64) error {
	s.seek[file] = offset
	return nil
}

func (s *SeekMem) Find(file string) (int64, error) {
	v := s.seek[file]
	return v, nil
}

func NewSeekMem() *SeekMem {
	return &SeekMem{
		seek: make(map[string]int64),
	}
}
