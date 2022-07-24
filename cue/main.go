package main

import (
	"cuelang.org/go/cue"
	"encoding/json"
	"fmt"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/common"
	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1alpha1"
)

type EnvConfig struct {
	Gender    string                  `json:"gender"`
	Placement []v1alpha1.EnvPlacement `json:"placement"`
	Selector  *v1alpha1.EnvSelector   `json:"selector"`
	Fourier   Fourier                 `json:"fourier"`
}

type Fourier struct {
	Fourier01 string `json:"fourier01"`
	Fourier02 string `json:"fourier02"`
}

type ab struct{ A, B int }

func main() {
	envConfig := EnvConfig{
		Gender: "fourier-flink-target01",
		Placement: []v1alpha1.EnvPlacement{
			v1alpha1.EnvPlacement{
				ClusterSelector: &common.ClusterSelector{
					Name: "cluster-flink",
				},
				NamespaceSelector: &v1alpha1.NamespaceSelector{
					Name: "fourier-flink-target01",
				},
			},
		},
		Selector: &v1alpha1.EnvSelector{
			[]string{"fourierapp05-fouriercomponent-01"},
		},
		Fourier: Fourier{
			Fourier01: "f01",
			Fourier02: "f02",
		},
	}
	envConfigMarshal, err := json.Marshal(envConfig)
	if err != nil {
		panic(err)
	}

	var r cue.Runtime
	envConfigInstance, _ := r.Compile("envConfig", envConfigMarshal)
	envConfigValue := envConfigInstance.Value()

	// 查找结构体中的字符串
	fourier := envConfigValue.Lookup("fourier", "fourier01")
	if fourier.Exists() {
		fourierStr, err := fourier.String()
		//bytes, err := fourier.MarshalJSON()
		if err != nil {
			panic(err)
		}
		fmt.Println(fourierStr)
	}

	// 查找结构体中的slice
	placement := envConfigValue.Lookup("placement")
	if placement.Exists() {
		// 方式一
		marshalJSON, err := placement.MarshalJSON()
		if err != nil {
			panic(err)
		}
		fmt.Println(marshalJSON)

		// 方式二
		placements := []v1alpha1.EnvPlacement{}
		err = placement.Decode(&placements)
		if err != nil {
			panic(err)
		}
		fmt.Println(placements)

	}

	// 设置cue模板的值，应该是只能设置变量的值
	root, err := r.Compile("test", `
	#Provider: {
		ID: string
		notConcrete: bool
	}
	`)
	if err != nil {
		panic(err)
	}
	spec := root.LookupDef("#Provider")
	providerInstance := spec.Fill("12345", "ID")
	print01, err := providerInstance.Eval().MarshalJSON()
	if err != nil {
		fmt.Println(print01)
	}
	root, err = root.Fill(providerInstance, "providers", "myprovider")
	if err != nil {
		panic(err)
	}
	print02, err := root.Value().MarshalJSON()
	if err != nil {
		fmt.Println(print02)
	}
	got := fmt.Sprint(root.Value())

	if got != `{#Provider: C{ID: string, notConcrete: bool}, providers: {myprovider: C{ID: (string & "12345"), notConcrete: bool}}}` {
		panic(got)
	}
}
