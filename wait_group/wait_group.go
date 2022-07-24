package wait_group

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// waitgroup会允许go先到GMP的g，g在p的队列进行等待而不分配m，从而实现waitgroup的等待，则不是直接不允许进入g
func Wait() {
	runtime.GOMAXPROCS(1)
	wg := sync.WaitGroup{}
	str := []string{"a", "b", "c"}
	for _, s := range str {
		wg.Add(1)
		go func() {
			time.Sleep(5 * time.Second)
			fmt.Println(s)
		}()
		wg.Done()
	}
	time.Sleep(20 * time.Second)
}
