package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"time"
)

func main() {
	group, ctx := errgroup.WithContext(context.Background())
	for _, v := range []int{5, 10, 15} {
		tempI := v
		group.Go(func() error {
			return func01(ctx, tempI)
		})
	}
	if err := group.Wait(); err != nil {
		logrus.Info("出错结束:", err.Error())
	}
	time.Sleep(time.Second * 20)
}

func func01(ctx context.Context, i int) error {
	c := make(chan error, 1)
	go func() {
		time.Sleep(time.Duration(i) * time.Second)
		err := func02(i)
		c <- err
	}()

	select {
	case <-ctx.Done():
		logrus.Info("ctx done")
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func func02(i int) error {
	logrus.Info("运行 ", i)
	return errors.New(string(i))
}
