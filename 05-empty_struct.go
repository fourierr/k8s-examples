package main

import (
	"fmt"
	"sync"
	"unsafe"
)

// 空结构体和零长数组（两个复合类型）都仅仅是一个占位符，不占用空间，这里编译器进行了优化，如果结构体或数组的unsafe.sizeof=0则直接返回zerobase

// NoUnkeyedLiterals 必须用key来初始化结构体
type NoUnkeyedLiterals struct{}

// DoNotCompare 不允许结构体比较
type DoNotCompare [0]func()

// DoNotCopy 不允许结构体拷贝、值传递
type DoNotCopy [0]sync.Mutex

type User struct {
	// 必须用key来初始化结构体
	NoUnkeyedLiterals
	// 不允许结构体比较
	DoNotCompare
	// 不允许结构体拷贝、值传递
	DoNotCopy
	Age     int
	Address string
}

func main() {
	u := User{Age: 21, Address: "beijing"}
	// todo 试一下值传递会怎么样
	test(u)

	fmt.Println(unsafe.Sizeof(NoUnkeyedLiterals{}))
	fmt.Printf("%p\n", &DoNotCopy{})
	fmt.Printf("%p\n", &NoUnkeyedLiterals{})
	fmt.Printf("%p\n", &DoNotCompare{})
}

func test(user User) {
	fmt.Printf("%+v", user)
}
