package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

func main() {

	log.Println("Starting...")

	// UTC是世界统一时间，time.FixedZone是利用了UTC的偏移
	secondsEastOfUTC := int((8 * time.Hour).Seconds())
	// 定义一个cron运行器
	c := cron.New(cron.WithSeconds(),
		cron.WithLocation(time.FixedZone("UTC-8", secondsEastOfUTC)))

	// 定时5秒，每5秒执行print5
	c.AddFunc("*/5 * * * * *", func() {
		log.Println("Run 5s cron")
	})

	// 开始
	c.Start()
	defer c.Stop()

	// 可以在running之后，继续添加定时任务
	c.AddFunc("*/7 * * * * *", func() {
		log.Println("Run 7s cron")
	})
	//inspect(cache.Entries())

	// 这是一个使用time包实现的定时器，与cron做对比
	//t1 := time.NewTimer(time.Second * 10)

	for {
		select {}
	}
}
