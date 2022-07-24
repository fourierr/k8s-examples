package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

func getFlag() {
	fmt.Println("111")
}
func main() {

	stopCh := make(chan struct{})
	for i := 0; i < 5; i++ {
		// 每隔10s运行一次getFlag函数
		go wait.Until(getFlag, 10*time.Second, stopCh)
	}
	<-stopCh
}
