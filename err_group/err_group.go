package err_group

import (
	"fmt"
	"sync"
	"time"
)

func TestErrGroup4s(paramString string,group *sync.WaitGroup) {
	defer group.Done()
	time.Sleep(3 * time.Second)
	fmt.Println("4s param：", paramString)
	time.Sleep(3 * time.Second)
}

func TestErrGroup2s(paramString string) error {
	time.Sleep(2 * time.Second)
	fmt.Println("2s param：", paramString)
	time.Sleep(2 * time.Second)
	return nil
}

func TestErrGroup3s(paramString string) error{
	time.Sleep(3 * time.Second)
	fmt.Println("3s param：", paramString)
	time.Sleep(3 * time.Second)
	return fmt.Errorf("3s出错了")
}