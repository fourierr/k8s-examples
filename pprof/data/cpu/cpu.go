package cpu

import "github.com/sirupsen/logrus"

type CPU struct {
}

func (c *CPU) Name() string {
	return "string"
}

func (c *CPU) Run() {
	logrus.Info(c.Name(), "Run")
	for i := 0; i < 100000000; i++ {

	}
}
