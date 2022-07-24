package deep_copy

type Struct1 struct {
	Name  string
	Phone string
}

//type Struct2 struct {
//	Name  string
//	Phone string
//}

func GetSlice1() []Struct1 {
	slice1 := []Struct1{}
	slice1 = append(slice1, Struct1{"name1", "phone1"})
	slice1 = append(slice1, Struct1{"name2", "phone2"})
	return slice1
}

func GetSlice2() []Struct1 {
	slice2 := []Struct1{}
	slice2 = append(slice2, Struct1{"name2", "phone2"})
	slice2 = append(slice2, Struct1{"name1", "phone1"})

	return slice2
}
