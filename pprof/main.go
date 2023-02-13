package main

import (
	"k8s-examples/pprof/data"
	"k8s-examples/pprof/data/block"
	"k8s-examples/pprof/data/cpu"
	"k8s-examples/pprof/data/goroutine"
	"k8s-examples/pprof/data/mem"
	"k8s-examples/pprof/data/mutex"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"
)

var cmds = []data.Cmd{
	&cpu.CPU{},
	&mem.Mem{},
	&goroutine.Goroutine{},
	&block.Block{},
	&mutex.Mutex{},
}

func main() {
	// 对阻塞操作的追踪
	runtime.SetBlockProfileRate(1)
	// 对死锁的追踪
	runtime.SetMutexProfileFraction(1)

	go func() {
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}()

	for {
		for _, cmd := range cmds {
			cmd.Run()
		}
		time.Sleep(time.Second)
	}
}
