package main

import (
	"fmt"
	"github.com/google/go-querystring/query"
	"github.com/kyokomi/emoji"
)

func main() {
	// 打印表情符号
	fmt.Println("Hello World Emoji!")
	emoji.Println(":beer: Beer!!!")

	pizzaMessage := emoji.Sprint("I like a :pizza: and :sushi:!!")
	fmt.Println(pizzaMessage)

	// 将结构体快速拼接成url的QueryParam
	// 该包的优点就是简单、快速。只要定义一个结构体调用该包的Encode函数就能将结构体中的字段自动拼接成url的查询参数。其缺点就是性能差
	opt := Options{"foo", true, 2, []int{1, 2, 3}}
	v, _ := query.Values(opt)

	// will output: "q=foo&all=true&page=2&score=1,2,3"
	// 际输出时会有url的转义
	fmt.Print(v.Encode())
}

type Options struct {
	Query   string `url:"q"`
	ShowAll bool   `url:"all"`
	Page    int    `url:"page"`
	// 通过在tag中增加comma标签，代表以逗号将值进行连接（实际输出时会有url的转义）
	Score []int `url:"score,comma"`
}
