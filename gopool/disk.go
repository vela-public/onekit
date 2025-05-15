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
	q := define[[]byte](ctx, options...)
	disk := q.option.Disk
	q.queue = &DiskQueue{
		dq: diskqueue.NewWithDiskSpace(disk.Name, disk.Path,
			disk.MaxBytesDiskSpace, disk.MaxBytesPerFile,
			disk.MinMsgSize, disk.MaxMsgSize,
			disk.SyncEvery, disk.SyncTimeout, disk.ErrHandle),
	}

	for i := 0; i < q.option.Workers; i++ {
		q.workers[i] = q.NewWorker(i)
		q.mapping[i] = q.NewExdata()
	}

	go q.supervise()
	return q
}
