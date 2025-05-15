package gopool

type Packet[T any] struct {
	Data T
	w    *Worker[T]
}

func (p *Packet[T]) Exdata() any {
	return p.w.Exdata()
}

func (p *Packet[T]) WorkerID() int {
	return p.w.id
}

func (p *Packet[T]) Cancel() {
	p.w.cancel()
}

func (p *Packet[T]) Errorf(format string, v ...any) {
	p.w.errorf(format, v...)
}
