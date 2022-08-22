package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/uuid"
)

func main() {
	// 生成uuid
	uid := uuid.NewUUID()
	fmt.Println(uid)

	// 生成随机数
	randInt := rand.IntnRange(0, 10)
	fmt.Println(randInt)
}
