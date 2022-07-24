package main

import (
	"fmt"
	"sync"
)

func main() {
	var scene sync.Map

	// 将键值对保存到sync.Map
	scene.Store("greece", 97)
	scene.Store("london", 100)
	scene.Store("egypt", 200)
	// 读取或写入,存在就读取，不存在就写入
	fmt.Println(scene.LoadOrStore("egyptt", 300))
	//scene.Store("egypt", 200)

	// 从sync.Map中根据键取值
	fmt.Println(scene.Load("london"))

	// 根据键删除对应的键值对
	scene.Delete("london")

	// 遍历所有sync.Map中的键值对
	scene.Range(func(k, v interface{}) bool {

		fmt.Println("iterate:", k, v)
		return true
	})
}
