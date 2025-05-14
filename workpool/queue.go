package workpool

import (
	"context"
	"fmt"
	"github.com/vela-public/go-diskqueue"
	"github.com/vela-public/onekit/libkit"
	"sync/atomic"
	"time"
)

const (
	UnDefine WorkerFlag = 0
	Running  WorkerFlag = 1 << iota
	Stop
	Panic
)

type QueueFSM struct {
	cnt int32
}

func (qf *QueueFSM) add(step int32) {
	atomic.AddInt32(&qf.cnt, step)
}

func (qf *QueueFSM) done() {
	atomic.AddInt32(&qf.cnt, -1)
}

type WorkerFlag uint8

type Worker[T any] struct {
	id      int
	flag    WorkerFlag
	context context.Context
	cancel  context.CancelFunc
	todo    func(T) error
	queue   QueueLine[T]
	errorf  func(format string, v ...any)
}

func (w *Worker[T]) run(fsm *QueueFSM) {
	defer func() {
		if e := recover(); e != nil {
			w.errorf("%v\n%s", e, libkit.StackTrace[string](1024, false))
			w.flag = Panic
		}
		w.cancel()
		fsm.done()
	}()

	fsm.add(1)
	w.flag = Running

	for {
		select {
		case <-w.context.Done():
			w.flag = Stop
			w.errorf("queue.%d worker exit", w.id)
			return
		case t, ok := <-w.queue.ReadChan():
			if !ok {
				return
			}
			err := w.todo(t)
			if err != nil {
				w.errorf("%v", err)
			}
		}
	}

}

type Queue[T any] struct {
	option  *Option
	fsm     *QueueFSM
	queue   QueueLine[T]
	context context.Context
	cancel  context.CancelFunc
	workers []*Worker[T]
	error   func(error)
	todo    func(T) error
}

func (q *Queue[T]) errorf(format string, v ...any) {
	err := fmt.Errorf(format, v...)
	if q.error != nil {
		q.error(err)
	}
}

func (q *Queue[T]) NewWorker(id int, fsm *QueueFSM) *Worker[T] {
	ctx, cancel := context.WithCancel(q.context)

	w := &Worker[T]{
		id:      id,
		flag:    UnDefine,
		context: ctx,
		cancel:  cancel,
		errorf:  q.errorf,
		queue:   q.queue,
	}

	w.todo = func(t T) error { return q.todo(t) }
	go w.run(fsm)
	return w
}

func (q *Queue[T]) Alive() int {
	return int(q.fsm.cnt)
}

func (q *Queue[T]) ping(t time.Time) {
	sz := len(q.workers)
	for i := 0; i < sz; i++ {
		w := q.workers[i]
		switch w.flag {
		case Panic:
			q.workers[i] = q.NewWorker(i, q.fsm)
			q.errorf("queue.%d restart", i)
		case UnDefine:
			q.workers[i] = q.NewWorker(i, q.fsm)
			q.errorf("queue.%d start", i)
		case Stop:
			return
		case Running:
			continue
		}
	}
}

func (q *Queue[T]) Stop() {
	q.cancel()
}

func (q *Queue[T]) Push(data T) {
	select {
	case <-q.context.Done():
		return
	default:
		err := q.queue.Push(data)
		if err != nil {
			q.option.Disk.ErrHandle(diskqueue.ERROR, "queue push data %v", err)

		}
	}
}

func (q *Queue[T]) Handler(fn func(T)) {
	q.todo = func(t T) error {
		fn(t)
		return nil
	}
}

type HandlerType[T any] interface {
	Do(T) error
}

func (q *Queue[T]) HandlerOf(obj interface{ Do(T) error }) {
	q.todo = func(t T) error {
		return obj.Do(t)
	}
}

func (q *Queue[T]) HandlerE(fn func(T) error) {
	q.todo = fn
}

func (q *Queue[T]) SetErrHandler(fn func(error)) {
	q.error = fn
}

func (q *Queue[T]) closeAll() {
	sz := len(q.workers)
	for i := 0; i < sz; i++ {
		w := q.workers[i]
		w.cancel()
	}
}

func (q *Queue[T]) supervise() {
	ticker := time.NewTicker(time.Duration(q.option.Tick) * time.Second)
	defer func() {
		ticker.Stop()
		q.queue.Close()
		q.closeAll()
	}()

	for {
		select {
		case <-q.context.Done():
			q.errorf("queue ticker exit")
			return
		case t := <-ticker.C:
			q.ping(t)
		}
	}
}

func NewQueue[T any](ctx context.Context, options ...func(*Option)) *Queue[T] {
	opt := &Option{
		Workers: 32,
		Cache:   0,
		Tick:    5,
	}

	for _, fn := range options {
		fn(opt)
	}

	q := &Queue[T]{
		option: opt,
		fsm:    &QueueFSM{},
	}
	q.context, q.cancel = context.WithCancel(ctx)

	if opt.Cache > 0 {
		q.queue = NewChanQueue[T](opt.Cache)
	} else {
		q.queue = NewChanQueue[T](opt.Cache)
	}

	q.workers = make([]*Worker[T], opt.Workers)
	for i := 0; i < opt.Workers; i++ {
		q.workers[i] = q.NewWorker(i, q.fsm)
	}

	go q.supervise()
	return q
}
