package main

import "fmt"

/*
	Functional Options函数选项模式（简称FOP模式）
	既保持了兼容性，而且每增加1个新属性只需要1个With函数即可，大大减少了修改代码的风险
*/

type Student struct {
	Name string
	Age  int
	Sex  string
}

//定义类型函数
type StudentOption func(*Student)

//创建带有age的构造函数
func WithAge(age int) StudentOption {
	return func(s *Student) {
		s.Age = age
	}
}

func WithSex(sex string) StudentOption {
	return func(s *Student) {
		s.Sex = sex
	}
}

//创建带有默认值的构造函数
func NewStudent(name string, options ...StudentOption) *Student {
	student := &Student{Name: name}
	for _, o := range options {
		o(student)
	}
	return student
}

func main() {
	student := NewStudent("fourier", WithAge(6), WithSex("男"))
	fmt.Println(student)
	teacher := NewTeacher("fourier", WithTeacherAge(20), WithGender("男"))
	fmt.Println(teacher)
}

type Teacher struct {
	Name   string
	Age    int
	Gender string
}

type TeacherOptions func(*Teacher)

func WithTeacherAge(age int) TeacherOptions {
	return func(teacher *Teacher) {
		teacher.Age = age
	}
}

func WithGender(gender string) TeacherOptions {
	return func(teacher *Teacher) {
		teacher.Gender = gender
	}
}

func NewTeacher(name string, options ...TeacherOptions) *Teacher {
	teacher := &Teacher{Name: name}
	for _, option := range options {
		option(teacher)
	}
	return teacher
}
