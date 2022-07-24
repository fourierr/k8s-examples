package main

import (
	"fmt"
	"proj_test/env_config"
)

func main() {
	env_config.LoadEnvironment()
	regionList := env_config.GetEnv().MngCluster.ApiServer
	fmt.Println(regionList)
	for _, region := range regionList {
		fmt.Println(region)
	}
}
