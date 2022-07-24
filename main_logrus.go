package main

import (
	"github.com/sirupsen/logrus"
)

func init() {
	logr.SetFormatter(&logrus.JSONFormatter{})
	// 添加自己实现的Hook
	logr.AddHook(&DefaultFieldsHook{})
	// 设置日志打印级别
	logr.SetLevel(logrus.InfoLevel)
}
var logr = logrus.New()

func main() {

	logr.WithFields(logrus.Fields{
		"field_author":"fourier",
	}).Warnf("infoMessage")

	logr.Warnf("errorMessage")
}

type DefaultFieldsHook struct {
}

func (df *DefaultFieldsHook) Fire(entry *logrus.Entry) error {
	entry.Data["author"] = "fourier"
	return nil
}

func (df *DefaultFieldsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}