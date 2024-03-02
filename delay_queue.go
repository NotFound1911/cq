package cq

import (
	"context"
	"fmt"
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
		dequeueSignal: make(chan struct{}, c),
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
	d.q.Enqueue(val)
	d.mutex.Unlock()
	d.enqueueSignal <- struct{}{}
	return nil
}

// Out 方法从DelayQueue中尝试取出一个元素。
// 如果元素的延迟时间到达，则返回该元素。
// 如果在等待过程中上下文被取消，则返回错误。
func (d *DelayQueue[T]) Out(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}
	// 定义一个计时器，用于等待元素的延迟时间
	var timer *time.Timer
	for {
		d.mutex.Lock()
		first, err := d.q.Peek()
		d.mutex.Unlock()
		switch err {
		case nil:
			// 拿到了元素
			// 获取该元素的延迟时间
			delay := first.Delay()
			// 如果延迟时间小于等于0，说明该元素不需要等待，可以直接取出
			if delay <= 0 {
				d.mutex.Lock()
				first, err := d.q.Peek() // 二次确认
				if err != nil {
					d.mutex.Unlock()
					continue
				}
				if first.Delay() <= 0 {
					first, err = d.q.Dequeue() // 如果确实小于等于0，则从队列中取出该元素
					d.mutex.Unlock()
					// 出队
					d.dequeueSignal <- struct{}{}
					return first, err
				}
				d.mutex.Unlock()
			}
			if timer == nil {
				timer = time.NewTimer(delay)
			} else {
				timer.Stop()
				timer.Reset(delay)
			}
			select {
			case <-timer.C:
				// 元素到期 进入下一个循环
			case <-d.enqueueSignal:
				// 来了新元素 进入下一个循环
				fmt.Println("enqueueSignal")
			case <-ctx.Done():
				var t T
				return t, ctx.Err()
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
