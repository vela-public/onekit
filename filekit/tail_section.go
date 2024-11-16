package filekit

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type SeekInfo struct {
	Offset int64 `lua:"offset"`
	Whence int   `lua:"whence"`
}

type Section struct {
	flag     ErrNo
	info     error
	tail     *FileTail
	path     string
	seek     int64
	location SeekInfo
	file     *os.File
	buff     *bufio.Reader
	time     time.Time //start time
}

func (s *Section) open() (stop bool) {
	if !s.follow() {
		s.flag = Paused
		return true
	}

	file, err := OpenFile(s.path)
	if err != nil {
		s.info = err
		return !os.IsNotExist(err)
	}

	ret, err := file.Seek(s.seek, io.SeekStart)
	if err != nil {
		s.tail.E("%s seek fail %v", s.path, err)
	} else {
		s.tail.E("%s seek follow %d", s.path, ret)
	}

	s.file = file
	s.buff = bufio.NewReaderSize(file, s.tail.setting.Buffer)
	s.time = time.Now()
	go s.line()
	s.flag = Running
	return true
}

func (s *Section) follow() bool {
	seek := s.tail.SeekTo(s.path)
	stat, err := os.Stat(s.path)
	if err != nil {
		s.flag = Paused
		s.info = err
		s.tail.E("%s stat fail %v", s.path, err)
		return false
	}

	size := stat.Size()
	if size == seek {
		s.flag = Paused
		s.info = io.EOF
		s.seek = size
		s.tail.logger.Infof("%s not change offset:%d size:%d", s.path, seek, size)
		return false
	}

	s.tail.E("%s record offset:%d size:%d", s.path, seek, size)
	if seek > size {
		s.seek = 0
	} else {
		s.seek = seek
	}

	return true
}

func (s *Section) Handle(raw []byte) {
	sz := len(raw)
	if sz <= 1 {
		return
	}

	if raw[sz-1] == s.tail.setting.Delim {
		raw = raw[:sz-1]
		sz -= 1
	}

	v := &Line{
		File: s.path,
		Text: raw,
		Size: sz,
	}

	s.tail.Input(v)
}

func (s *Section) close() {

	if s.file == nil {
		s.tail.E("%s not file handle", s.path)
		return
	}

	seek, e := s.file.Seek(0, io.SeekCurrent)
	if e != nil {
		s.tail.E("%s current seek error %v", s.path, e)
		return
	}

	s.tail.Tell(s.path, seek)

	err := s.file.Close()
	if err != nil {
		s.tail.E("%s fd close fail %v", s.path, err)
		return
	}

	s.tail.E("Tx %s fd close succeed", s.path)
}

func (s *Section) detect() {
	switch s.flag {
	case Nothing:
		s.start()
	case Paused:
		s.reload()
	case Stopped:
		s.tail.logger.Errorf("%s %v", s.path, s.info)
	case Cleaned:
		s.tail.logger.Errorf("%s clean", s.path)
	case Done:
		s.tail.logger.Errorf("%s done", s.path)
	case Running:
		//todo
	default:

	}
}

func (s *Section) start() {
	s.tail.WaitFor(func() (stop bool) {
		stop = s.open()
		return
	})
}

func (s *Section) reload() {
	s.tail.WaitFor(func() (stop bool) {
		stop = s.open()
		return
	})
}

func (s *Section) line() {
	defer func() {
		if r := recover(); r != nil {
			s.flag = Paused
			buff := make([]byte, 1024*32)
			runtime.Stack(buff, false)
			s.tail.E("file:%s error:%v stack:\n%s", s.path, r, string(buff))
		}
		s.close()
	}()

	for {

		select {
		case <-s.tail.Done():
			s.flag = Done
			s.tail.E("%s readline exit", s.path)
			return

		default:
			s.tail.Wait()

			raw, err := s.buff.ReadBytes(s.tail.setting.Delim)
			if err == nil {
				s.Handle(raw)
				continue
			}

			switch err.Error() {
			case io.EOF.Error():
				s.flag = Paused
				s.Handle(raw)
				return

			case os.ErrClosed.Error():
				s.flag = Paused
				s.Handle(raw)
				return
			default:
				s.flag = Paused
				s.info = err
				s.tail.E("%s read line error %v", s.path, err)
				return
			}
		}
	}
}

func (ft *FileTail) Section(path string) *Section {
	path, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		path = filepath.Clean(path)
	}

	return &Section{
		path: path,
	}
}
