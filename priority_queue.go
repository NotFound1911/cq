package cq

import (
	"errors"
)

var (
	ErrOutOfCapacity = errors.New("ekit: 超出最大容量限制")
	ErrEmptyQueue    = errors.New("空队列")
)

type PriorityQueue[T any] struct {
	compare Comparator[T]
	cap     int
	data    []T
}

func (p *PriorityQueue[T]) Len() int {
	return len(p.data) - 1
}

// Cap 无界队列返回0，有界队列返回创建队列时设置的值
func (p PriorityQueue[T]) Cap() int {
	return p.cap
}

func (p *PriorityQueue[T]) IsBoundless() bool {
	return p.cap <= 0
}

func (p *PriorityQueue[T]) isFull() bool {
	return p.cap > 0 && len(p.data)-1 == p.cap
}

func (p *PriorityQueue[T]) isEmpty() bool {
	return len(p.data) == 1
}

func (p *PriorityQueue[T]) Peek() (T, error) {
	if p.isEmpty() {
		var t T
		return t, ErrEmptyQueue
	}
	return p.data[1], nil
}

// heapifyUp 向上移动
func (p *PriorityQueue[T]) heapifyUp(id int) {
	for id > 1 {
		if p.compare(p.data[id], p.data[id/2]) < 0 { // 如果更小
			p.data[id], p.data[id/2] = p.data[id/2], p.data[id] // 交换移动
			id /= 2
		} else {
			break
		}
	}
}

// heapifyDown 向下移动
func (p *PriorityQueue[T]) heapifyDown(id int) {
	child := id * 2
	for child < len(p.data) {
		other := id*2 + 1                                                           // 另外一个还在
		if (other < len(p.data)) && (p.compare(p.data[other], p.data[child]) < 0) { // 选择另外一个较小的孩子交换
			child = other
		}
		// 维护较小的
		if p.compare(p.data[child], p.data[id]) < 0 {
			p.data[child], p.data[id] = p.data[id], p.data[child]
			id = child
			child = id * 2
		} else {
			break
		}
	}
}

func (p *PriorityQueue[T]) Enqueue(t T) error {
	if p.isFull() {
		return ErrOutOfCapacity
	}
	p.data = append(p.data, t)   // 插入末尾
	p.heapifyUp(len(p.data) - 1) // 最后一个元素向上移动
	return nil
}

func (p *PriorityQueue[T]) Dequeue() (T, error) {
	if p.isEmpty() {
		var t T
		return t, ErrEmptyQueue
	}
	pop := p.data[1]
	// 将最小元素放入末尾
	p.data[1], p.data[len(p.data)-1] = p.data[len(p.data)-1], p.data[1]
	p.data = p.data[:len(p.data)-1]
	p.heapifyDown(1)
	return pop, nil
}

// NewPriorityQueue 创建优先队列 capacity <= 0 时，为无界队列，否则有有界队列
func NewPriorityQueue[T any](cap int, compare Comparator[T]) *PriorityQueue[T] {
	sliceCap := cap + 1
	if cap < 1 {
		cap = 0
		sliceCap = 32
	}
	return &PriorityQueue[T]{
		cap:     cap,
		data:    make([]T, 1, sliceCap),
		compare: compare,
	}
}
