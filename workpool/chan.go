package workpool

type ChanQueue[T any] struct {
	ch chan T
}

func NewChanQueue[T any](size int) *ChanQueue[T] {
	if size == 0 {
		return &ChanQueue[T]{
			ch: make(chan T),
		}
	}

	return &ChanQueue[T]{
		ch: make(chan T, size),
	}
}
func (c *ChanQueue[T]) Close() {
	close(c.ch)
}

func (c *ChanQueue[T]) Pop() (T, bool) {
	v, ok := <-c.ch
	return v, ok
}

func (c *ChanQueue[T]) ReadChan() <-chan T {
	return c.ch
}

func (c *ChanQueue[T]) Push(v T) error {
	c.ch <- v
	return nil
}
