package main

import (
	"fmt"
	"os"
	"runtime/pprof"
)

// pprof_init go tool pprof -http=:8888 cpu.prof
// go tool pprof cpu.prof
func pprof_init(run func()) {
	fc, err := os.Create("./cpu.prof")
	if err != nil {
		fmt.Println("create cpu.prof err:", err.Error())
		return
	}
	defer fc.Close()

	// 开始分析cpu
	err = pprof.StartCPUProfile(fc)
	if err == nil {
		defer pprof.StopCPUProfile()
	}
	// --- 内存 分析示例 start---
	fm, err := os.Create("./memory.prof")
	if err != nil {
		fmt.Println("create memory.prof err:", err.Error())
		return
	}
	defer fm.Close()
	// 开始分析内存
	err = pprof.WriteHeapProfile(fm)
	if err != nil {
		fmt.Println("write heap prof err:", err.Error())
		return
	}
	run()
}
