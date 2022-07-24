package gjson_practise

import (
	"fmt"
	"github.com/tidwall/gjson"
)

const json = `{"name":[{"first":"Janet","last":"Prichard"}, {"abc": 1, "e" : 2}],"age":47}`

var jsonInterface interface{}

func main() {
	// 1、把string转result，再取数据
	gjson.Parse(json).Get("name")
	// 2、把[]byte转result，再取数据
	gjson.ParseBytes([]byte(json)).Get("name")
	// 3、从string格式的json直接取数据
	v := gjson.Get(json, "name")
	// 4、interface{}要先强制转换为[]byte或string，再用gjson的三种方法（单层不如直接强转为map，多层考虑用gjson）
	gjson.Parse(jsonInterface.(string)).Get("name")
	gjson.ParseBytes(jsonInterface.([]byte)).Get("name")
	gjson.Get(jsonInterface.(string), "name")

	if v.String() == "" {
		fmt.Println("nil")
	} else {
		fmt.Println(v.String())
		fmt.Println(v.Type)
	}


}
