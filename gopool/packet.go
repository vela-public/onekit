package gopool

type Packet[T any] struct {
	Data T
	flag int // 标志位 0:content 1:timer 2:canceler
	w    *Worker[T]
}

func (p *Packet[T]) Flag() int {
	return p.flag
}
func (p *Packet[T]) Timer() bool {
	return p.flag == 1
}
func (p *Packet[T]) Canceler() bool {
	return p.flag == 2
}
func (p *Packet[T]) Content() bool {
	return p.flag == 0
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
