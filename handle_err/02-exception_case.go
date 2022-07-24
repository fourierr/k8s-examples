package handle_err

import "fmt"

// 对于不应该出现的分支，建议直接使用panic
func fun1() {
	switch s := getStr(); s {
	case "A":
	// ...
	case "B":
	// ...
	case "C":
	// ... 
	case "D":
	// ...
	default:
		panic(fmt.Sprintf("invalid str %v", s))
	}
}
func getStr() string {
	return "E"
}
