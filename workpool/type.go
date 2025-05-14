package workpool

type QueueLine[T any] interface {
	Pop() (T, bool)
	Push(T) error
	ReadChan() <-chan T
	Close()
}
