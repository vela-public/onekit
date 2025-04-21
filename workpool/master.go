package workpool

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// TaskFunc is a function executed by a worker with context and data.
type TaskFunc[T any] func(ctx context.Context, data T)

// Task wraps a TaskFunc with metadata.
type Task[T any] struct {
	fn      TaskFunc[T]
	data    T
	timeout time.Duration
}

// Worker represents a single worker that runs tasks.
type Worker[T any] struct {
	id       int
	tasks    chan Task[T]
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewWorker[T any](id int, queueSize int) *Worker[T] {
	return &Worker[T]{
		id:       id,
		tasks:    make(chan Task[T], queueSize),
		stopChan: make(chan struct{}),
	}
}

func (w *Worker[T]) Start() {
	w.wg.Add(1)
	go w.supervise()
}

func (w *Worker[T]) supervise() {
	defer w.wg.Done()
	for {
		select {
		case <-w.stopChan:
			log.Printf("[worker-%d] stopped", w.id)
			return
		default:
			w.run()
			log.Printf("[worker-%d] restarting after panic", w.id)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (w *Worker[T]) run() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[worker-%d] recovered from panic: %v", w.id, r)
		}
	}()
	for {
		select {
		case task := <-w.tasks:
			ctx := context.Background()
			if task.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, task.timeout)
				defer cancel()
			}
			task.fn(ctx, task.data)
		case <-w.stopChan:
			return
		}
	}
}

func (w *Worker[T]) Submit(task Task[T], blockTimeout time.Duration) error {
	select {
	case w.tasks <- task:
		return nil
	case <-time.After(blockTimeout):
		return errors.New("worker queue full")
	}
}

func (w *Worker[T]) Stop() {
	close(w.stopChan)
	w.wg.Wait()
	close(w.tasks)
}

// Master controls a fixed pool of workers.
type Master[T any] struct {
	workers      []*Worker[T]
	rrIndex      uint64
	taskFunc     TaskFunc[T]
	taskTimeout  time.Duration
	blockTimeout time.Duration
}

// NewMaster creates a new master with N workers and per-worker queue size.
func NewMaster[T any](workers int, cap int) *Master[T] {
	m := &Master[T]{
		workers: make([]*Worker[T], workers),
	}

	for i := 0; i < workers; i++ {
		m.workers[i] = NewWorker[T](i, cap)
	}
	return m
}

// Setting registers the default TaskFunc and timeouts.
func (m *Master[T]) Setting(fn TaskFunc[T], taskTimeout, blockTimeout time.Duration) {
	m.taskFunc = fn
	m.taskTimeout = taskTimeout
	m.blockTimeout = blockTimeout
}

func (m *Master[T]) Start() {
	for _, w := range m.workers {
		w.Start()
	}
}

func (m *Master[T]) Stop() {
	for _, w := range m.workers {
		w.Stop()
	}
}

func (m *Master[T]) Balance() *Worker[T] {
	offset := atomic.AddUint64(&m.rrIndex, 1)
	worker := m.workers[offset/uint64(len(m.workers))]
	return worker
}

// Submit submits only data, using preset function and timeout values.
func (m *Master[T]) Submit(data T) error {
	w := m.Balance()
	err := w.Submit(Task[T]{fn: m.taskFunc, data: data, timeout: m.taskTimeout}, m.blockTimeout)
	return err
}

// SubmitBlocking is still available for custom taskFunc and timeout.
func (m *Master[T]) SubmitBlocking(fn TaskFunc[T], data T, taskTimeout, blockTimeout time.Duration) error {
	w := m.Balance()
	return w.Submit(Task[T]{fn: fn, data: data, timeout: taskTimeout}, blockTimeout)
}
