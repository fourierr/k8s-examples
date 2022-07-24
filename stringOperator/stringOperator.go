package stringOperator

import (
	"fmt"
	"unicode"
)

func TravelString() {
	str := "hello傅里叶"
	// byte类型
	for i := 0; i < len(str); i++ {
		fmt.Printf("%v(%c) ", str[i], str[i])
	}
	fmt.Println()
	// rune类型
	for _, s := range str {
		fmt.Printf("%v(%c) ", s, s)
	}
	fmt.Println()
}

func UpdateString() {
	str1 := "hello"
	byteStr := []byte(str1) // 强制类型转换
	byteStr[0] = 'a'
	fmt.Println(string(byteStr))

	str2 := "傅里叶"
	runeStr := []rune(str2) // 强制类型转换
	runeStr[0] = '佛'
	fmt.Println(string(runeStr))
}
func CountHanString() {
	str := "hello傅里叶"
	var count int
	for _, s := range str {
		if unicode.Is(unicode.Han, s) {
			count++
		}
	}
	fmt.Println(count)
}
