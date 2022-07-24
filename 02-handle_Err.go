package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
)

func main() {

	//logrus.Error(fun1())
	//logrus.Info(strings.Split(err.Error(), ": "))
	//logrus.Info(errors.Cause(err))

}

func writeConfig(data interface{}) (err error) {
	defer func() {
		if err != nil {
			err = errors.WithMessage(err, "writing configuration")
		}
	}()
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile("config.json", b, 0644); err != nil {
		return err
	}
	return
}

func fun1() error {
	return errors.WithMessage(fun2(), "1111")
}

func fun2() error {
	return errors.WithMessage(fun3(), "2222")
}

func fun3() error {
	return errors.WithMessage(fun4(), "3333")
	//return nil
}

func fun4() error {
	return errors.New("4444")
	//return nil
}
func CopyFile(src, dst string) error {
	r, err := os.Open(src)
	if err != nil { // 错误检查
		return fmt.Errorf("copy %s %s: %v", src, dst, err) // 错误处理
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil { // 错误检查
		return fmt.Errorf("copy %s %s: %v", src, dst, err) // 错误处理
	}

	if _, err := io.Copy(w, r); err != nil { // 错误检查
		// 错误处理
		w.Close()
		os.Remove(dst)
		return fmt.Errorf("copy %s %s: %v", src, dst, err)
	}

	if err := w.Close(); err != nil { // 错误检查
		// 错误处理
		os.Remove(dst)
		return fmt.Errorf("copy %s %s: %v", src, dst, err)
	}
	return nil
}
