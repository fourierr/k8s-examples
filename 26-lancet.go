package main

import (
	"fmt"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
)

// labels

func main() {
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
		"e": 5,
	}
	isEven := func(_ string, value int) bool {
		return value%2 == 0
	}

	result := maputil.Filter(m, isEven)

	//result := maputil.Filter(m, isEven)

	fmt.Println(result)
	// Output:
	// map[b:2 d:4]

	// -------------------------------------------
	nums := []int{1, 2, 3, 4, 5}

	isEven2 := func(i, num int) bool {
		return num%2 == 0
	}

	result2 := slice.CountBy(nums, isEven2)

	fmt.Println(result2)

	// Output:
	// 2
}
