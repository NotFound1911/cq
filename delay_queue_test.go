package cq

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type delayElemTest struct {
	val int
	end time.Time
}

func (d delayElemTest) Delay() time.Duration {
	return d.end.Sub(time.Now())
}

func TestDelayQueue_InAndOut(t *testing.T) {
	dq := NewDelayQueue[delayElemTest](5)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		err := dq.In(ctx, delayElemTest{
			end: time.Now().Add(time.Second * time.Duration(i)),
			val: i,
		})
		require.NoError(t, err)
	}
	for i := 0; i < 5; i++ {
		element, err := dq.Out(ctx)
		assert.NoError(t, err)
		assert.Equal(t, i, element.val)

	}
}
func TestDelayQueue_In_Full(t *testing.T) {
	dq := NewDelayQueue[delayElemTest](5)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		err := dq.In(ctx, delayElemTest{
			end: time.Now().Add(time.Second * time.Duration(i)),
			val: i,
		})
		require.NoError(t, err)
	}
	flag := false
	go func() {
		time.Sleep(time.Second)
		assert.Equal(t, flag, false)
		// 阻塞
		err := dq.In(ctx, delayElemTest{
			end: time.Now().Add(time.Second * 1),
			val: 1,
		})
		require.NoError(t, err)
		assert.Equal(t, flag, true)
	}()
	time.Sleep(time.Second * 2)
	flag = true
	_, err := dq.Out(ctx)
	assert.NoError(t, err)
}

func TestDelayElem_Out_Empty(t *testing.T) {
	dq := NewDelayQueue[delayElemTest](5)
	ctx := context.Background()
	flag := false
	go func() {
		time.Sleep(time.Second)
		assert.Equal(t, flag, false)
		// 阻塞
		element, err := dq.Out(ctx)
		assert.NoError(t, err)
		require.NoError(t, err)
		assert.Equal(t, element.val, 10)
		assert.Equal(t, flag, true)
	}()
	time.Sleep(time.Second * 2)
	flag = true
	err := dq.In(ctx, delayElemTest{
		end: time.Now().Add(time.Second),
		val: 10,
	})
	require.NoError(t, err)
}
