package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

func main() {
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	//stopCh := make(chan struct{})

	// 每隔10s运行一次getFlag函数
	wait.Until(func() {
		fmt.Println("i")
	}, 3*time.Second, timeout.Done())

	time.Sleep(3 * time.Second)
	//stopCh <- struct{}{}

	time.Sleep(100 * time.Second)

}
