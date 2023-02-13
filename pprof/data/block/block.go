package block

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Block struct {
}

func (b *Block) Name() string {
	return "Block"
}

func (b *Block) Run() {
	logrus.Info(b.Name(), "Run")
	<-time.After(1 * time.Second)
}
