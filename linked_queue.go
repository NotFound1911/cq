package cq

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type LinkedQueue[T any] struct {
	// *node[T]
	head unsafe.Pointer // 头节点的指针
	// *node[T]
	tail unsafe.Pointer // 尾节点的指针
}
type node[T any] struct {
	data T
	// *node[T]
	next unsafe.Pointer // 指向下一个节点的指针
}

func (l *LinkedQueue[T]) In(val T) error {
	newNode := &node[T]{data: val}
	newPtr := unsafe.Pointer(newNode)
	for {
		tailPtr := atomic.LoadPointer(&l.tail) // 原子地加载尾节点的指针
		tail := (*node[T])(tailPtr)
		tailNext := atomic.LoadPointer(&tail.next) // 原子地加载尾节点的下一个节点指针
		if tailNext != nil {
			continue // 被并发修改了
		}
		// 使用CompareAndSwapPointer确保原子地设置尾节点的下一个节点为新节点
		// 如果成功，则更新尾节点为新节点
		// 先指向新节点，再调整tail节点
		if atomic.CompareAndSwapPointer(&tail.next, tailNext, newPtr) {
			atomic.CompareAndSwapPointer(&l.tail, tailPtr, newPtr)
			return nil
		}
	}
}

func (l *LinkedQueue[T]) Out() (T, error) {
	for {
		headPtr := atomic.LoadPointer(&l.head)
		head := (*node[T])(headPtr)
		tailPtr := atomic.LoadPointer(&l.tail)
		tail := (*node[T])(tailPtr)
		if head == tail { // 如果头节点和尾节点相同，则队列为空
			var t T
			return t, fmt.Errorf("empty queue")
		}
		// 原子地加载头节点的下一个节点指针
		headNextPtr := atomic.LoadPointer(&head.next)
		if atomic.CompareAndSwapPointer(&l.head, headPtr, headNextPtr) {
			headNext := (*node[T])(headNextPtr)
			return headNext.data, nil
		}
	}
}
func NewLinkedQueue[T any]() *LinkedQueue[T] {
	n := &node[T]{
		next: unsafe.Pointer(nil),
	}
	ptr := unsafe.Pointer(n)

	return &LinkedQueue[T]{
		head: ptr,
		tail: ptr,
	}
}
