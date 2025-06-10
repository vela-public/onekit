package datalog

import (
	"context"
	"fmt"
	"github.com/vela-public/onekit/gopool"
	"sync"
	"time"
)

var (
	public = new(struct {
		enable bool
		ctx    context.Context
		cancel context.CancelFunc
		once   sync.Once
		queue  *gopool.Queue[Event]
	})
)

type Event struct {
	Time    time.Time
	Level   string
	Key     string
	Message string
}

func Enable(thread int, console bool, handler func(Event)) {
	public.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		public.enable = true
		public.ctx = ctx
		public.cancel = cancel
		public.queue = gopool.NewQueue[Event](ctx, gopool.Workers(thread))
		public.queue.With(func(pkt *gopool.Packet[Event]) error {
			handler(pkt.Data)
			if console {
				fmt.Printf("%s %s %s %s\n", pkt.Data.Time.Format("2006-01-02 15:04:05.000"), pkt.Data.Level, pkt.Data.Key, pkt.Data.Message)
			}
			return nil
		})
	})
}

func Err(key string) func(format string, v ...any) {
	return func(format string, v ...any) {
		if !public.enable {
			return
		}
		public.queue.Push(Event{
			Time:    time.Now(),
			Level:   "error",
			Key:     key,
			Message: fmt.Sprintf(format, v...),
		})
	}
}

func Info(key string) func(format string, v ...any) {
	return func(format string, v ...any) {
		if !public.enable {
			return
		}
		public.queue.Push(Event{
			Time:    time.Now(),
			Level:   "info",
			Key:     key,
			Message: fmt.Sprintf(format, v...),
		})
	}
}
