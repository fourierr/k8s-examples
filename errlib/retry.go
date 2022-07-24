package errlib

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

func NewDefaultRetryConf() wait.Backoff {
	return wait.Backoff{
		Steps:    5,
		Duration: time.Duration(2) * time.Second,
		Factor:   1.0,
		Jitter:   0.1,
	}
}
func RetryOnErr(bacoff wait.Backoff, fn func() error) error {
	return OnErr(bacoff, IsError, fn)
}

func IsError(err error) bool {
	return err != nil
}

func OnErr(backoff wait.Backoff, retriable func(error) bool, fn func() error) error {
	var lastConflictErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		err := fn()
		switch {
		case err == nil:
			return true, nil
		case retriable(err):
			lastConflictErr = err
			return false, nil
		default:
			return false, err
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastConflictErr
	}
	return err
}
