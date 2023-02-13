package mem

import (
	"github.com/sirupsen/logrus"
	"k8s-examples/pprof/constants"
)

type Mem struct {
	// 声明一个切片，每个元素是一兆的数组
	buffer [][constants.Mi]byte
}

func (m *Mem) Name() string {
	return "Mem"
}

func (m *Mem) Run() {
	logrus.Info(m.Name(), "Run")
	max := constants.Gi
	for len(m.buffer)*constants.Mi < max {
		m.buffer = append(m.buffer, [constants.Mi]byte{})
	}
}
