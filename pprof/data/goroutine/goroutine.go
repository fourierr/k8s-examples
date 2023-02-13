package goroutine

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Goroutine struct {
}

func (g *Goroutine) Name() string {
	return "Goroutine"
}

func (g *Goroutine) Run() {
	logrus.Info(g.Name(), "Run")
	for i := 0; i < 10; i++ {
		go func() {
			time.Sleep(30 * time.Second)
		}()
	}
}
