package worker

import "fmt"

/*
	实现select的优先级
*/
func Worker(ch1, ch2 <-chan string, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case job1 := <-ch1:
			fmt.Println(job1)
		case job2 := <-ch2:
		Priority:
			for {
				select {
				case job1 := <-ch1:
					fmt.Println(job1)
				default:
					break Priority
				}
			}
			fmt.Println(job2)
		}
	}
}
