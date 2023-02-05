package mutex

import (
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Mutex struct {
}

func (m *Mutex) Name() string {
	return "Mutex"
}

func (m *Mutex) Run() {
	logrus.Info(m.Name(), "Run")
	mutex := &sync.Mutex{}
	mutex.Lock()
	go func() {
		time.Sleep(time.Second)
		mutex.Unlock()
	}()
	mutex.Lock()
}
