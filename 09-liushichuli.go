package main

import "fmt"

func main() {

	multiply := func(
		done <-chan interface{},
		intStream <-chan int,
		multiplier int,
		err error,
	) (<-chan int, error) {
		if err != nil {
			return nil, err
		}
		multipliedStream := make(chan int)
		go func() {
			defer close(multipliedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case multipliedStream <- i * multiplier:
				}
			}
		}()
		return multipliedStream, nil
	}

	add := func(
		done <-chan interface{},
		intStream <-chan int,
		additive int,
		err error,
	) (<-chan int, error) {
		if err != nil {
			return nil, err
		}
		addedStream := make(chan int)
		go func() {
			defer close(addedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case addedStream <- i + additive:
				}
			}
		}()
		return addedStream, nil
	}

	gen := func(
		done <-chan interface{},
		integers ...int,
	) (<-chan int, error) {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for _, i := range integers {
				select {
				case <-done:
					return
				case intStream <- i:
				}
			}
		}()
		return intStream, nil
	}

	done := make(chan interface{})
	defer close(done)
	var err error
	intStream, err := gen(done, 1, 2, 3, 4)
	addStream, err := add(done, intStream, 1, err)
	pipeline, err := multiply(done, addStream, 2, err)
	for v := range pipeline {
		fmt.Println(v)
	}
}
