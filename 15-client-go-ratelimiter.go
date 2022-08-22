package main

import (
	"k8s.io/client-go/util/flowcontrol"
	ctrl "sigs.k8s.io/controller-runtime"
)

// 工作队列是 k8s.io/client-go/util/workqueue
// 令牌桶算法限速 flowcontrol.NewTokenBucketRateLimiter(qps float32, burst int)
// 排队指数限速 workqueue.NewItemExponentialFailureRateLimiter(baseDelay time.Duration, maxDelay time.Duration)
// baseDelay 基础限速时间，maxDelay 最大限速时间
// 计数器模式  workqueue.NewItemFastSlowRateLimiter(fastDelay, slowDelay time.Duration, maxFastAttempts int)
//fastDelay 快限速时间，slowDelay 慢限速时间，maxFastAttempts 快限速元素个数
// 混合模式(多种限速算法同时使用) NewMaxOfRateLimiter

// 默认是 k8s.io/client-go/util/flowcontrol
func main() {
	config := ctrl.GetConfigOrDie()
	config.QPS = 5
	config.Burst = 10
	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(5, 10)
}
