package main

import (
	"fmt"
	"proj_test/deep_copy"
	"reflect"
)

func main() {
	slice1 := deep_copy.GetSlice1()
	slice2 := deep_copy.GetSlice2()
	equal := reflect.DeepEqual(slice1, slice2)
	fmt.Println(equal)
}
