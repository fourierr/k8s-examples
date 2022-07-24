package inherit

import "fmt"

type Animal struct {
	Name    string `json:"name"`
	Feature string `json:"feature"`
}


func NewAnimal(name, feature string) *Animal{
	return &Animal{
		name,
		feature,
	}
}

func (a *Animal) Move(name string) {
	fmt.Printf("%s animal move \n", name)
}

func (a *Animal) Shout(name string) {
	fmt.Printf("%s animal shout \n", name)
}

type Dog struct {
	Name string `json:"name"`
	*Animal
}

func (d Dog) Move(name string)  {
	fmt.Printf("%s dog move \n", name)
}