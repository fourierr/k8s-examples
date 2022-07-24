package main

import (
	"encoding/json"
	"fmt"
)

const personJson = `{"age":47}`

//const personJson = `{"name":{"first":"Janet","last":"Prichard"},"age":47}`

type Person struct {
	Name Name `json:"name,omitempty"`
	Age  int  `json:"age,omitempty"`
}
type Name struct {
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
}

func main() {
	// method one----gjson_practise
	//parse_json := gjson_practise.Parse(personJson)
	//first1 := parse_json.Get("name").Get("first").String()
	//age1 := parse_json.Get("age").Int()
	//fmt.Println("first: " + first1)
	//fmt.Printf("age: %v", age1)
	//fmt.Println()

	// method two---- unmarshal
	//person_byte:=[]byte(personJson)
	//person := new(Person)
	//_ = json.Unmarshal(person_byte, person)
	//fmt.Println("first: ", person.Name.First)
	//fmt.Printf("age: %v", person.Age)
	//fmt.Println()

	// method three
	//person_marshal, _ := json.Marshal(person_json)
	//personMap := interface{}(person_marshal).(map[string]interface{})
	//nameMap := interface{}(personMap["name"]).(map[string]interface{})
	//age3 := personMap["age"].(int64)
	//first3 := nameMap["first"].(string)
	//fmt.Println("first: " + first3)
	//fmt.Printf("age: %v", age3)

	person_byte := []byte(personJson)
	person := new(Person)
	_ = json.Unmarshal(person_byte, person)
	if person.Age == 0 {
		fmt.Println(person.Age)
	}
	if person.Name.Last == "" {
		fmt.Println(person.Name.Last)
	}

	if person.Name == (Name{}) {
		fmt.Println(person.Name)
	}
}
