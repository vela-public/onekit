package filekit

import (
	"bufio"
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

type SeekInfo struct {
	Offset int64 `lua:"offset"`
	Whence int   `lua:"whence"`
}

type Section struct {
	again bool
	flag  ErrNo
	info  error
	tail  *FileTail
	path  string
	seek  int64
	file  *os.File
	time  time.Time //start time
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

	var ret int64
	if !s.again && s.tail.setting.Location.Whence != 0 {
		ret, err = file.Seek(s.tail.setting.Location.Offset, s.tail.setting.Location.Whence)
	} else {
		ret, err = file.Seek(s.seek, io.SeekStart)
	}

	s.again = true

	if err != nil {
		s.tail.Errorf("%s seek fail %v", s.path, err)
	} else {
		s.tail.Debugf("%s seek follow %d", s.path, ret)
	}

	s.file = file
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
		s.tail.Errorf("%s stat fail %v", s.path, err)
		return false
	}

	size := stat.Size()
	if size == seek {
		s.flag = Paused
		s.info = io.EOF
		s.seek = size
		s.tail.Debugf("%s not change offset:%d size:%d", s.path, seek, size)
		return false
	}

	s.tail.Debugf("%s record offset:%d size:%d", s.path, seek, size)
	if seek > size {
		s.seek = 0
	} else {
		s.seek = seek
	}

	return true
}

func (s *Section) Handle(raw string) {
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
		Text: []byte(raw),
		Size: sz,
	}

	s.tail.Input(v)
}

func (s *Section) SaveSeek() {
	if s.file == nil {
		s.tail.Errorf("%s not file handle", s.path)
		return
	}

	seek, e := s.file.Seek(0, io.SeekCurrent)
	if e != nil {
		s.tail.Errorf("%s current seek error %v", s.path, e)
		return
	}
	s.tail.Tell(s.path, seek)
}

func (s *Section) close() {

	if s.file == nil {
		s.tail.Errorf("%s not file handle", s.path)
		return
	}

	seek, e := s.file.Seek(0, io.SeekCurrent)
	if e != nil {
		s.tail.Errorf("%s current seek error %v", s.path, e)
		return
	}

	s.tail.Tell(s.path, seek)

	err := s.file.Close()
	if err != nil {
		s.tail.Errorf("%s fd close fail %v", s.path, err)
		return
	}

	s.tail.Debugf("Tx %s fd close succeed", s.path)
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
			s.tail.Errorf("file:%s error:%v stack:\n%s", s.path, r, string(buff))
		}
		s.close()
	}()

	reader := bufio.NewReader(s.file)
	var cnt uint32

	for {

		select {
		case <-s.tail.Done():
			s.flag = Done
			s.tail.Errorf("%s readline exit", s.path)
			return

		default:
			fsm := LineFSM{
				tail:    s.tail,
				scanner: reader,
				next:    false,
				err:     nil,
			}

			text, err := fsm.Read()
			if err == nil {
				s.Handle(text)
				continue
			}

			if atomic.AddUint32(&cnt, uint32(1)) > uint32(200) {
				s.SaveSeek()
			}

			switch err.Error() {
			case io.EOF.Error():
				s.flag = Paused
				s.Handle(text)
				return

			case os.ErrClosed.Error():
				s.flag = Paused
				s.Handle(text)
				return
			default:
				s.flag = Paused
				s.info = err
				s.tail.Errorf("%s read line error %v", s.path, err)
				return
			}
		}
	}
}
