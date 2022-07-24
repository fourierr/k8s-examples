package main

import "proj_test/inherit"

func main() {
	dog:=inherit.Dog{
		Name: "small dog",
		Animal:inherit.NewAnimal("small animal","animal feature"),
	}
	dog.Move("small dog")
	dog.Animal.Move("animal")
	dog.Shout("small dog")
}
