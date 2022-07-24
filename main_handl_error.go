package main

import "fmt"

func main() {
	fmt.Println("panic example")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic recover: %s", r)
		}
	}()
	panic("this is a panic")
}
