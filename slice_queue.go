package cq

import (
	"context"
	"golang.org/x/sync/semaphore"
	"sync"
)

// SliceQueue 基于切片的队列
type SliceQueue[T any] struct {
	data  []T
	head  int // 头
	tail  int // 尾巴
	count int // 队列中元素数量

	mutex   *sync.RWMutex
	enqueue *semaphore.Weighted // 用于控制入队操作的信号量
	dequeue *semaphore.Weighted // 用于控制出队操作的信号量
}

func NewSliceQueue[T any](cap int) *SliceQueue[T] {
	mu := &sync.RWMutex{}
	res := &SliceQueue[T]{
		data:    make([]T, cap),
		mutex:   mu,
		enqueue: semaphore.NewWeighted(int64(cap)),
		dequeue: semaphore.NewWeighted(int64(cap)),
	}
	// 相当于说，先出队的时候（完全没有入队过），必然阻塞
	_ = res.dequeue.Acquire(context.TODO(), int64(cap))
	return res
}
func (q *SliceQueue[T]) In(ctx context.Context, v T) error {
	// 尝试获取一个入队信号量，如果获取失败则阻塞
	err := q.enqueue.Acquire(ctx, 1)
	if err != nil {
		return err
	}
	// 到了这里，就相当于已经预留了一个座位
	q.mutex.Lock()
	defer q.mutex.Unlock()
	// 检查是否在上下文中发生了错误
	if ctx.Err() != nil {
		// 释放之前获取的入队信号量
		q.enqueue.Release(1)
		return ctx.Err()
	}
	// 元素放到尾部
	q.data[q.tail] = v
	q.tail++
	q.count++
	if q.tail == cap(q.data) {
		q.tail = 0
	}
	// 通知等待出队操作的 goroutine
	q.dequeue.Release(1)
	return nil
}

// Out ctx 用于超时控制，要么在超时内返回一个数据，要么返回一个 error
func (q *SliceQueue[T]) Out(ctx context.Context) (T, error) {
	var t T
	err := q.dequeue.Acquire(ctx, 1)
	if err != nil {
		return t, ctx.Err()
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if ctx.Err() != nil {
		// 释放之前获取的出队信号量
		q.dequeue.Release(1)
		return t, ctx.Err()
	}
	front := q.data[q.head]
	q.data[q.head] = t

	q.head++
	q.count--

	if q.head == cap(q.data) {
		q.head = 0
	}
	// 拿走一个元素, 就唤醒对面在等待空位的人
	q.enqueue.Release(1)
	return front, nil
}
func (q *SliceQueue[T]) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.isEmpty()
}

func (q *SliceQueue[T]) isEmpty() bool {
	return q.count == 0
}

func (q *SliceQueue[T]) isFull() bool {
	return q.count == cap(q.data)
}
