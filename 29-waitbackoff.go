package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/clock"
	"time"
)

func main() {
	start := time.Now()
	reflectorClock := clock.RealClock{}
	stopCh := make(<-chan struct{}, 1)
	backoffManager := wait.NewExponentialBackoffManager(800*time.Millisecond, 30*time.Second, 2*time.Minute, 2.0, 1.0, reflectorClock)

	wait.BackoffUntil(func() {
		fmt.Println("aaaa", time.Now())
	}, backoffManager, true, stopCh)
	fmt.Println("结束", time.Since(start))
}
