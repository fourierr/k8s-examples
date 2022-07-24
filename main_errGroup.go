package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"proj_test/err_group"
	"sync"
)

func main() {
	//获取的是值
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)
	go err_group.TestErrGroup4s("4s", &waitGroup)
	go err_group.TestErrGroup4s("又是4s", &waitGroup)
	waitGroup.Wait()
	fmt.Println("waitgroup结束")

	// 获取的是引用地址
	errGroup := new(errgroup.Group)
	errGroup.Go(func() error {
		return err_group.TestErrGroup2s("errgroup2s")
	})
	errGroup.Go(func() error {
		if err := err_group.TestErrGroup3s("errgroup3s"); err != nil {
			logrus.Errorf("出错了:%v",err)
			return err
		}
		return nil
	})
	errGroup.Wait()
	fmt.Println("errgroup结束")

}
