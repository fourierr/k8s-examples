package main

import "fmt"

/*
	go get golang.org/x/tools/cmd/stringer
	在同级目录下终端输入go generate就会自动生成code_string.go
*/
type ErrCode int

//go:generate stringer -type ErrCode -linecomment -output code_string.go
const (
	ERR_OK      ErrCode = iota //请求OK
	ERR_PARAMS                 //请求参数出错
	ERR_TIMEOUT                //请求超时
)

func main() {
	fmt.Println(ERR_OK)
	fmt.Println(ERR_PARAMS)
	fmt.Println(ERR_TIMEOUT)

	if ERR_OK == 0 {
		fmt.Println(ERR_OK)
	}

	if ERR_TIMEOUT == 2 {
		fmt.Println(ERR_TIMEOUT)
	}
}
