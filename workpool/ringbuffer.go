package workpool

import "sync"

type RingBuffer[T any] struct {
	buffer []T
	size   int
	head   int
	tail   int
	lock   sync.Mutex
	cond   *sync.Cond
}

func NewRingBuffer[T any](size int) *RingBuffer[T] {
	rb := &RingBuffer[T]{
		buffer: make([]T, size),
		size:   size,
		head:   0,
		tail:   0,
	}
	rb.cond = sync.NewCond(&rb.lock) // Create a new condition variable
	return rb
}

func (rb *RingBuffer[T]) Push(item T) bool {
	rb.lock.Lock()
	defer rb.lock.Unlock()

	if (rb.tail+1)%rb.size == rb.head {
		// Buffer is full, cannot push new item
		return false
	}

	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size
	rb.cond.Signal() //通知工作人员有新任务
	return true
}

func (rb *RingBuffer[T]) Pop() (T, bool) {
	rb.lock.Lock()
	defer rb.lock.Unlock()

	for rb.head == rb.tail {
		// Buffer is empty, wait until a task is available
		rb.cond.Wait()
	}

	item := rb.buffer[rb.head]
	rb.head = (rb.head + 1) % rb.size
	return item, true
}
