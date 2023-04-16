package main

import (
	"fmt"
	"github.com/VictoriaMetrics/fastcache"
	jsoniter "github.com/json-iterator/go"
)

type DataStruct struct {
	Uid  int    `json:"uid"`
	Name string `json:"name"`
}

var cache2 = fastcache.New(200 * 1024 * 1024)

func main() {

	var key = "aaa"
	var dataStruct = &DataStruct{
		Uid:  111,
		Name: "222",
	}
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	marshal, err := json.Marshal(dataStruct)
	if err != nil {
		return
	}
	cache2.Set([]byte(key), marshal)

	get := cache2.Get(nil, []byte(key))
	newDataStruct := &DataStruct{}
	err = json.Unmarshal(get, newDataStruct)
	if err != nil {
		return
	}
	fmt.Println(newDataStruct)
	if cache2.Has([]byte(key)) {
		cache2.Del([]byte(key))
	}
	getRetry := cache2.Get(nil, []byte(key))
	fmt.Println(getRetry)

}
