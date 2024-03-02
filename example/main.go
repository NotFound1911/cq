package main

import (
	"context"
	"fmt"
	"github.com/NotFound1911/cq"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync"
	"time"
)

type delayElem struct {
	val int
	end time.Time
}

func (d delayElem) Delay() time.Duration {
	return d.end.Sub(time.Now())
}
func runQueue() {
	runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪，block
	runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪，mutex
	go func() {
		// http://localhost:8080/debug/pprof/
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()
	dq := cq.NewDelayQueue[delayElem](20)
	ctx := context.Background()
	wg := sync.WaitGroup{}
	length := 1000000
	wg.Add(length)
	for i := 0; i < length; i++ {
		go func(val int) {
			t := time.Second * time.Duration(rand.Intn(10))
			err := dq.In(ctx, delayElem{
				val: i,
				end: time.Now().Add(t),
			})
			if err != nil {
				fmt.Printf("in queued:%v failed,err:%v\n", val, err)
			}
		}(i)
	}
	for i := 0; i < length; i++ {
		go func() {
			defer wg.Done()
			val, err := dq.Out(ctx)
			if err != nil {
				fmt.Printf("out queued:%+v failed,err:%v\n", val, err)
			}
		}()
	}
	wg.Wait()
	fmt.Println("finished")
	for {

	}
}

func main() {
	pprof_init(runQueue)
}
