package gopool

import (
	"context"
	"fmt"
	"github.com/vela-public/go-diskqueue"
	"github.com/vela-public/onekit/libkit"
	"golang.org/x/time/rate"
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
	ref     *Queue[T]
	id      int
	flag    WorkerFlag
	context context.Context
	cancel  context.CancelFunc
	queue   QueueLine[T]
	errorf  func(format string, v ...any)
}

func (w *Worker[T]) Exdata() any {
	return w.ref.mapping[w.id]
}

func (w *Worker[T]) handler(v T) error {
	if w.ref.private.Handler == nil {
		return nil
	}
	w.ref.Wait()

	return w.ref.private.Handler(&Packet[T]{
		Data: v,
		w:    w,
	})
}

func (w *Worker[T]) run() {
	defer func() {
		if e := recover(); e != nil {
			w.errorf("%v\n%s", e, libkit.StackTrace[string](1024, false))
			w.flag = Panic
		}
		w.cancel()
		w.ref.fsm.done()
		w.ref.doAfter(&Packet[T]{
			w: w,
		})
	}()

	w.ref.fsm.add(1)
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
			err := w.handler(t)
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
	workers []*Worker[T]
	mapping []any

	private struct {
		Context context.Context
		Cancel  context.CancelFunc
		Error   func(error)
		Handler func(*Packet[T]) error
		After   func(*Packet[T])
		Limit   *rate.Limiter
	}
}

func (q *Queue[T]) errorf(format string, v ...any) {
	err := fmt.Errorf(format, v...)
	if q.private.Error != nil {
		q.private.Error(err)
	}
}

func (q *Queue[T]) NewExdata() any {
	if q.option.Exdata == nil {
		return nil
	}
	return q.option.Exdata()
}

func (q *Queue[T]) doAfter(packet *Packet[T]) {
	if q.private.After != nil {
		q.private.After(packet)
	}
}

func (q *Queue[T]) NewWorker(id int) *Worker[T] {
	ctx, cancel := context.WithCancel(q.Context())

	w := &Worker[T]{
		id:      id,
		ref:     q,
		flag:    UnDefine,
		context: ctx,
		cancel:  cancel,
		errorf:  q.errorf,
		queue:   q.queue,
	}

	go w.run()
	return w
}

func (q *Queue[T]) Alive() int {
	return int(q.fsm.cnt)
}

func (q *Queue[T]) Context() context.Context {
	return q.private.Context
}

func (q *Queue[T]) ping(t time.Time) {
	select {
	case <-q.Context().Done():
		return
	default:
		sz := len(q.workers)
		for i := 0; i < sz; i++ {
			w := q.workers[i]
			switch w.flag {
			case Panic:
				q.workers[i] = q.NewWorker(i)
				q.errorf("queue.%d restart", i)
			case UnDefine:
				q.workers[i] = q.NewWorker(i)
				q.errorf("queue.%d start", i)
			case Stop:
				q.workers[i] = q.NewWorker(i)
			case Running:
				continue
			}
		}
	}
}

func (q *Queue[T]) Stop() {
	q.private.Cancel()
}

func (q *Queue[T]) Push(data T) {
	select {
	case <-q.Context().Done():
		return
	default:
		err := q.queue.Push(data)
		if err != nil {
			q.option.Disk.ErrHandle(diskqueue.ERROR, "queue push data %v", err)

		}
	}
}

func (q *Queue[T]) HandlerFunc(fn func(packet *Packet[T])) {
	q.private.Handler = func(packet *Packet[T]) error {
		fn(packet)
		return nil
	}
}

type HandlerType[T any] interface {
	Do(T) error
}

func (q *Queue[T]) Handler(obj interface{ Do(packet *Packet[T]) error }) {
	q.private.Handler = func(packet *Packet[T]) error {
		return obj.Do(packet)
	}
}

func (q *Queue[T]) HandlerFuncE(fn func(packet *Packet[T]) error) {
	q.private.Handler = fn
}

func (q *Queue[T]) SetErrHandler(fn func(error)) {
	q.private.Error = fn
}

func (q *Queue[T]) Limit(n int) {
	if n > 0 {
		q.private.Limit = rate.NewLimiter(rate.Limit(n), n)
	}
}

func (q *Queue[T]) After(fn func(packet *Packet[T])) {
	q.private.After = fn
}

func (q *Queue[T]) Wait() {
	if q.private.Limit != nil {
		_ = q.private.Limit.Wait(q.Context())
	}
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

	//初始化worker
	for i := 0; i < q.option.Workers; i++ {
		q.workers[i] = q.NewWorker(i)
		q.mapping[i] = q.NewExdata()
	}

	for {
		select {
		case <-q.Context().Done():
			q.errorf("queue ticker exit")
			return
		case t := <-ticker.C:
			q.ping(t)
		}
	}
}

func define[T any](parent context.Context, options ...func(option *Option)) *Queue[T] {
	opt := &Option{
		Workers: 32,
		Cache:   0,
		Tick:    1,
	}

	for _, fn := range options {
		fn(opt)
	}

	ctx, cancel := context.WithCancel(parent)

	q := &Queue[T]{
		option: opt,
		fsm:    &QueueFSM{},
	}
	q.private.Context = ctx
	q.private.Cancel = cancel
	q.workers = make([]*Worker[T], opt.Workers)
	q.mapping = make([]any, opt.Workers)

	return q
}

func NewQueue[T any](parent context.Context, options ...func(*Option)) *Queue[T] {
	q := define[T](parent, options...)
	q.queue = NewChanQueue[T](q.option.Cache)

	for i := 0; i < q.option.Workers; i++ {
		q.workers[i] = q.NewWorker(i)
		q.mapping[i] = q.NewExdata()
	}

	go q.supervise()
	return q
}
