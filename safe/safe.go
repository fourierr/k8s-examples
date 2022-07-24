package safe

import "fmt"

func Go(f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("panic recover ", err)
			}
		}()
		f()
	}()
}
