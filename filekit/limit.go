package filekit

import (
	"context"
	"golang.org/x/time/rate"
)

type limit struct {
	ctx    context.Context
	cancel context.CancelFunc
	rate   *rate.Limiter
}

func NewLimit(parent context.Context, n int) *limit {
	if n <= 0 {
		return &limit{}
	}

	ctx, cancel := context.WithCancel(parent)
	if n <= 0 {
		return &limit{rate: nil, ctx: ctx, cancel: cancel}
	}
	return &limit{ctx, cancel, rate.NewLimiter(rate.Limit(n), n*2)}
}

func (l *limit) wait() {
	if l.rate == nil {
		return
	}
	_ = l.rate.Wait(l.ctx)
}
