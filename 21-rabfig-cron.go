package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"time"
)

// labels

func main() {
	c := cron.New(cron.WithSeconds())

	c.AddFunc("CRON_TZ=Asia/Shanghai 10 * * ? * 1,2,3,4,5", func() {
		fmt.Println(time.Now(), "执行时间")
	})
	c.Start()

	for {
		select {}
	}
}
