package gopool

import (
	"github.com/vela-public/go-diskqueue"
	"time"
)

type Option struct {
	Workers int
	Cache   int
	Tick    int
	Exdata  func() any
	Disk    struct {
		Name              string
		Path              string
		Error             bool
		MaxBytesDiskSpace int64
		MaxBytesPerFile   int64
		MinMsgSize        int32
		MaxMsgSize        int32
		SyncEvery         int64
		SyncTimeout       time.Duration
		ErrHandle         func(diskqueue.LogLevel, string, ...any)
	}
}

func Exdata(fn func() any) func(option *Option) {
	return func(opt *Option) {
		opt.Exdata = fn
	}
}

func Workers(n int) func(option *Option) {
	return func(opt *Option) {
		opt.Workers = n
	}
}

func Cache(n int) func(option *Option) {
	return func(opt *Option) {
		opt.Cache = n
	}
}

func Ticker(n int) func(option *Option) {
	return func(opt *Option) {
		opt.Tick = n
	}
}

func DiskSpace(name string, dataPath string,
	maxBytesDiskSpace int64, maxBytesPerFile int64,
	minMsgSize int32, maxMsgSize int32,
	syncEvery int64, syncTimeout time.Duration, logf func(diskqueue.LogLevel, string, ...any)) func(*Option) {
	return func(opt *Option) {
		opt.Disk.Name = name
		opt.Disk.Path = dataPath
		opt.Disk.MaxBytesDiskSpace = maxBytesDiskSpace
		opt.Disk.MaxBytesPerFile = maxBytesPerFile
		opt.Disk.MinMsgSize = minMsgSize
		opt.Disk.MaxMsgSize = maxMsgSize
		opt.Disk.SyncEvery = syncEvery
		opt.Disk.SyncTimeout = syncTimeout
		opt.Disk.ErrHandle = logf
	}
}
