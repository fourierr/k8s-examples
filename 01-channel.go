package main

import (
	"fmt"
	"time"
)

func main() {
	go considerChan()
	time.Sleep(time.Second * 100)
}

func considerChan() {
	ch := make(chan int)

	go func() {
		for i := 0; i < 15; i++ {
			fmt.Println("写前", i)
			ch <- i
			fmt.Println("写后", i)
			time.Sleep(time.Second)
			if i == 10 {
				close(ch)
				break
			}
		}
	}()


	for {
		item, ok := <-ch
		if ok {
			fmt.Println(item)
		} else {
			break
		}
	}
	for item := range ch {
		fmt.Println(item)
	}
	fmt.Println("这里技术")
	time.Sleep(time.Second * 3)
}
