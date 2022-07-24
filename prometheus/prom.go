package main

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	MyCounter   prometheus.Counter
	MyGauge     prometheus.Gauge
	MyHistogram prometheus.Histogram
	MySummary   prometheus.Summary
	MyCounter2  *prometheus.CounterVec
)

// init 注册指标
func init() {
	// 1.定义指标（类型，名字，帮助信息）
	MyCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "my_counter_total",
		Help: "自定义counter",
	})
	// 定义gauge类型指标
	MyGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "my_gauge_num",
		Help: "自定义gauge",
	})
	// 定义histogram
	MyHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "my_histogram_bucket",
		Help:    "自定义histogram",
		Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5}, // 需要指定桶
	})
	// 定义Summary
	MySummary = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "my_summary_bucket",
		Help: "自定义summary",
		// 这部分可以算好后在set
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	})
	// 定义带标签的指标
	MyCounter2 = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "my_counter_total",
			Help: "自定义counter",
		},
		// 标签集合
		[]string{"label1", "label2"},
	)

	// 2.注册指标
	prometheus.MustRegister(MyCounter)
	prometheus.MustRegister(MyGauge)
	prometheus.MustRegister(MyHistogram)
	prometheus.MustRegister(MySummary)
	prometheus.MustRegister(MyCounter2)
}

// Sayhello
func Sayhello(w http.ResponseWriter, r *http.Request) {
	// 接口请求量递增
	MyCounter.Inc()
	MyCounter2.With(prometheus.Labels{"label1": "1", "label2": "2"}).Inc()
	fmt.Fprintf(w, "Hello Wrold!")
}
