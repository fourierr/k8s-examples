package env_config

import (
	"fmt"
	"github.com/vrischmann/envconfig"
	"log"
	"reflect"
)

type Env struct {
	ExpireTime int `envconfig:"default=30"`
	Retry      struct {
		Limits   int `envconfig:"default=5"`
		Interval int `envconfig:"default=2"`
	}
	MngCluster struct {
		ApiServer  []string `envconfig:"default=csdev-biz-01-sk-mg;csdev-biz-01-qh-mg"`
		Controller string   `envconfig:"default=csdev-biz-01-sk-mg"`
	}
}

var env Env

func LoadEnvironment() {
	err := envconfig.Init(&env)
	if err != nil {
		log.Fatalln(err)
	} else {
		elements := reflect.ValueOf(&env).Elem()
		for i := 0; i < elements.NumField(); i++ {
			fmt.Printf("%s ---> %s\n", elements.Type().Field(i).Name, elements.Field(i).Interface())
		}
	}
}

func GetEnv() Env {
	return env
}