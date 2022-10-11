package main

import "fmt"

// Employee Factory Generator
/*
根据角色创建单独的工厂，然后可以使用这些工厂创建实例。（“从”中培养开发**developerFactory**人员）
注意，工厂本身的类型func是可以传递给其他函数的，本质上也是函数式编程。
由于工厂是 type func，很容易地被其他功能消费。
在这种方法中，工厂本身不能改变对象自身。例如developerFactory.AnnualSalary = 100是非法的。
*/
type Employee struct {
	Name, Designation string
	AnnualSalary      int
}

// NewEmployeeFactory functional approach
func NewEmployeeFactory(designation string, annualSalary int) func(name string) *Employee {
	return func(name string) *Employee {
		return &Employee{name, designation, annualSalary}
	}
}

func Factory_Generator1() {
	developerFactory := NewEmployeeFactory("developer", 60000)

	developer := developerFactory("Surya")
	fmt.Println("Developer:", developer)

	managerfactory := NewEmployeeFactory("manager", 80000)

	manager := managerfactory("John")
	fmt.Println("Manager:", manager)
}

// Structural approach
/*
创建了一个单独的工厂结构体EmployeeFactory。
创建一个构造函数EmployeeFactory。
将 a 定义为返回对象Create()的成员。EmployeeFactory Employee
这种方法的优点是工厂可以改变对象。例如bossFactory.AnnualIncome = 500是合法的。
因为EmployeeFactory它是一种特殊类型，与使用这种类型的最终用户不同，应该知道暴露developerFactory (line:20)了哪些功能（在这种情况下）。Create这可以使用接口来完成。
*/

type EmployeeFactory struct {
	Designation  string
	AnnualSalary int
}

func NewEmployeeFactory2(designation string, annualSalary int) *EmployeeFactory {
	return &EmployeeFactory{designation, annualSalary}
}

func (e *EmployeeFactory) Create(name string) *Employee {
	return &Employee{name, e.Designation, e.AnnualSalary}
}

func Factory_Generator2() {
	bossFactory := NewEmployeeFactory2("CEO", 10000000)
	boss := bossFactory.Create("Bill Gates")
	fmt.Println(boss)
}
