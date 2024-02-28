package cq

import (
	"context"
	"sync"
	"time"
)

type Delayable interface {
	// Delay 实时计算
	Delay() time.Duration
}
type DelayElem struct {
	end time.Time
}

func (d DelayElem) Delay() time.Duration {
	return d.end.Sub(time.Now())
}

type DelayQueue[T Delayable] struct {
	q             *PriorityQueue[T]
	mutex         *sync.Mutex
	enqueueSignal chan struct{}
	dequeueSignal chan struct{}
}

func NewDelayQueue[T Delayable](c int) *DelayQueue[T] {
	m := &sync.Mutex{}
	res := &DelayQueue[T]{
		mutex:         m,
		enqueueSignal: make(chan struct{}, c),
		q: NewPriorityQueue[T](c, func(src T, dst T) int {
			srcDelay := src.Delay()
			dstDelay := dst.Delay()
			if srcDelay > dstDelay {
				return 1
			}
			if srcDelay == dstDelay {
				return 0
			}
			return -1
		}),
	}
	return res
}
func (d *DelayQueue[T]) In(ctx context.Context, val T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	d.mutex.Lock()
	for d.q.isFull() {
		d.mutex.Unlock()
		select {
		case <-d.dequeueSignal:
			d.mutex.Lock()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	// 获取队列中最靠前的元素（即延迟时间最短的元素）
	first, err := d.q.Peek()
	if err != nil {
		d.mutex.Unlock()
		return err
	}
	d.q.Enqueue(val)
	d.mutex.Unlock()
	if val.Delay() < first.Delay() {
		close(d.enqueueSignal)
	}
	return nil
}
func (d *DelayQueue[T]) Out(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	var timer *time.Timer
	for {
		d.mutex.Lock()
		first, err := d.q.Peek()
		d.mutex.Unlock()
		switch err {
		case nil:
			// 拿到了元素
			delay := first.Delay()
			if delay <= 0 {

			}
		case ErrEmptyQueue:
			// 阻塞，等待新元素或报错
			select {
			case <-d.enqueueSignal:
			case <-ctx.Done():
				var t T
				return t, ctx.Err()
			}
		default:
			var t T
			return t, err
		}
	}
}
