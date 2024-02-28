package cq

import (
	"testing"
)

func TestLinkedQueue_InAndOut(t *testing.T) {
	q := NewLinkedQueue[int]()
	const numElements = 10

	// 向队列中添加元素
	for i := 0; i < numElements; i++ {
		if err := q.In(i); err != nil {
			t.Fatalf("Failed to enqueue element %d: %v", i, err)
		}
	}

	// 从队列中移除并验证元素
	for i := 0; i < numElements; i++ {
		val, err := q.Out()
		if err != nil {
			t.Fatalf("Failed to dequeue element %d: %v", i, err)
		}
		t.Log("val:", val)
		if val != i {
			t.Fatalf("Dequeued wrong value. Expected %d, got %d", i, val)
		}
	}

	// 尝试从空队列中移除元素，应返回错误
	_, err := q.Out()
	if err == nil {
		t.Fatal("Expected error when dequeuing from an empty queue")
	}
}
