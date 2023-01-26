package main

import (
	"encoding/json"
	"fmt"
)

type StructA struct {
	B StructB `json:"b.b"`
}
type StructB struct {
	C string `json:"c"`
}

func (in *StructA) MarshalJSON() ([]byte, error) {
	type Alias StructA
	tmp := &struct{ *Alias }{}
	tmp.Alias = (*Alias)(in)
	return json.Marshal(tmp.Alias)
}

func (in *StructA) UnmarshalJSON(src []byte) error {
	type Alias StructA
	tmp := &struct{ *Alias }{}
	if err := json.Unmarshal(src, tmp); err != nil {
		return err
	}
	(*StructA)(tmp.Alias).DeepCopyInto(in)
	return nil
}

func (in *StructA) DeepCopyInto(out *StructA) {
	*out = *in
}

func main() {
	structA := &StructA{B: StructB{
		C: "cccc",
	}}
	marshal, err := json.Marshal(structA)
	if err != nil {
		return
	}
	fmt.Println(marshal)
	newStructA := &StructA{}
	err = json.Unmarshal(marshal, newStructA)
	if err != nil {
		return
	}
	fmt.Println(newStructA)
}
