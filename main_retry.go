package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"proj_test/errlib"
)

func main() {
	times := 0
	retryErr := errlib.RetryOnErr(errlib.NewDefaultRetryConf(), func() error {
		times++
		return apply()
	})
	if retryErr != nil {
		logrus.Errorf("catch error")
	}
}

func apply() error {
	logrus.Infof("apply")
	// 编写业务逻辑
	return fmt.Errorf("error")
}
