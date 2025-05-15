package gopool

import (
	"context"
	"github.com/vela-public/go-diskqueue"
)

type DiskQueue struct {
	dq diskqueue.Interface
}

func (dk *DiskQueue) Pop() ([]byte, bool) {
	ch := dk.dq.ReadChan()
	data, ok := <-ch
	return data, ok
}

func (dk *DiskQueue) ReadChan() <-chan []byte {
	return dk.dq.ReadChan()
}
func (dk *DiskQueue) Push(data []byte) error {
	return dk.dq.Put(data)
}

func (dk *DiskQueue) Close() {
	_ = dk.dq.Close()
}

func NewDiskQueue(ctx context.Context, options ...func(*Option)) *Queue[[]byte] {
	opt := &Option{
		Workers: 32,
		Cache:   0,
		Tick:    1,
	}

	for _, fn := range options {
		fn(opt)
	}

	q := &Queue[[]byte]{
		option: opt,
		fsm:    &QueueFSM{},
	}
	q.private.Context, q.private.Cancel = context.WithCancel(ctx)

	dq := diskqueue.NewWithDiskSpace(opt.Disk.Name, opt.Disk.Path,
		opt.Disk.MaxBytesDiskSpace, opt.Disk.MaxBytesPerFile,
		opt.Disk.MinMsgSize, opt.Disk.MaxMsgSize,
		opt.Disk.SyncEvery, opt.Disk.SyncTimeout, opt.Disk.ErrHandle)

	q.queue = &DiskQueue{dq: dq}
	q.workers = make([]*Worker[[]byte], opt.Workers)
	q.mapping = make([]any, opt.Workers)
	for i := 0; i < opt.Workers; i++ {
		q.workers[i] = q.NewWorker(i)
		q.mapping[i] = q.NewExdata()
	}

	go q.supervise()
	return q
}
