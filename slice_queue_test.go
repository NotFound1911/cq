package cq

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	w := semaphore.NewWeighted(20)
	ch := make(chan any, 1)
	go func() {
		err := w.Acquire(context.Background(), 1)
		t.Log(err)
		ch <- err
	}()
	<-ch
}

func TestSliceQueue_In(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
		in   int
		q    *SliceQueue[int]

		wantErr  error
		wantData []int
	}{
		{
			name: "超时",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
				return ctx
			}(),
			in: 10,
			q: func() *SliceQueue[int] {
				q := NewSliceQueue[int](2)
				_ = q.In(context.Background(), 11)
				_ = q.In(context.Background(), 12)
				return q
			}(),
			wantErr:  context.DeadlineExceeded,
			wantData: []int{11, 12},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.q.In(tc.ctx, tc.in)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantData, tc.q.data)
		})
	}
}
func TestSliceQueue_Out(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
		q    *SliceQueue[int]

		wantErr  error
		wantData int
	}{
		{
			name: "超时",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
				return ctx
			}(),
			q: func() *SliceQueue[int] {
				q := NewSliceQueue[int](2)
				return q
			}(),
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.q.Out(tc.ctx)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantData, res)
		})
	}
}
