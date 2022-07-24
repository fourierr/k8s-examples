package three_dot

import "fmt"



func TestOne(args ...string)  {
	for _, v:= range args{
		fmt.Println(v)
	}
}

