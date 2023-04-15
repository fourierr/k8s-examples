package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime"
	"sync"
)

// labels

func main() {
	// 1、sync.Mutex允许在共享资源上互斥访问（不能同时访问）
	mutex := &sync.Mutex{}
	mutex.Lock()
	// Update 共享变量 (比如切片，结构体指针等)
	mutex.Unlock()

	// 2、sync.RWMutex读写互斥锁，互斥写，并发读，只有在频繁读取和不频繁写入的场景里，才应该使用sync.RWMutex
	rwMutex := &sync.RWMutex{}
	rwMutex.Lock()
	// Update 共享变量
	rwMutex.Unlock()

	rwMutex.RLock()
	// Read 共享变量
	rwMutex.RUnlock()

	//	3、sync.WaitGroup一个goroutine等待一组goroutine执行完成
	// 当计数器等于0时，则Wait()方法会立即返回。否则它将阻塞执行Wait()方法的goroutine直到计数器等于0时为止
	wg := &sync.WaitGroup{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			// Do something
			wg.Done()
		}()
	}
	wg.Wait()
	// 继续往下执行...

	// 4、sync.Map是一个并发版本的Go语言的map
	// 对map有频繁的读取和不频繁的写入时；多个goroutine读取，写入和覆盖不相交的键时。
	m := &sync.Map{}

	// 添加元素
	m.Store(1, "one")
	m.Store(2, "two")

	// 获取元素1
	value, contains := m.Load(1)
	if contains {
		fmt.Printf("%s\n", value.(string))
	}

	// 返回已存value，否则把指定的键值存储到map中
	value, loaded := m.LoadOrStore(3, "three")
	if !loaded {
		fmt.Printf("%s\n", value.(string))
	}

	m.Delete(3)

	// 迭代所有元素
	m.Range(func(key, value interface{}) bool {
		fmt.Printf("%d: %s\n", key.(int), value.(string))
		return true
	})

	// 5、sync.Pool并发池，负责安全地保存一组对象，想要重用共享的和长期存在的对象（例如，数据库连接）时
	// 写入缓冲区并将结果持久保存到文件中的函数示例
	pool := &sync.Pool{}

_:
	writeFile(pool, "file")

	// 6、sync.Once保证一个函数仅执行一次
	once := &sync.Once{}
	for i := 0; i < 4; i++ {
		i := i
		go func() {
			once.Do(func() {
				fmt.Printf("first %d\n", i)
			})
		}()
	}

	// 7、sync.Cond发出信号（一对一）或广播信号（一对多）到goroutine
	cond := sync.NewCond(&sync.Mutex{})
	s := make([]int, 1)
	for i := 0; i < runtime.NumCPU(); i++ {
		go printFirstElement(s, cond)
	}

	i := 666
	cond.L.Lock()
	s[0] = i
	// 会解除一个goroutine的阻塞状态
	cond.Signal()
	cond.L.Unlock()

	cond.L.Lock()
	s[0] = i
	// 会通知所有goroutine解除阻塞状态
	cond.Broadcast()
	cond.L.Unlock()
}

// 写入缓冲区并将结果持久保存到文件中的函数示例
func writeFile(pool *sync.Pool, filename string) error {
	buf := pool.Get().(*bytes.Buffer)
	defer pool.Put(buf)
	// Reset 缓存区，不然会连接上次调用时保存在缓存区里的字符串foo
	// 编程foofoo 以此类推
	buf.Reset()
	buf.WriteString("foo")
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

func printFirstElement(s []int, cond *sync.Cond) {
	cond.L.Lock()
	// cond.Wait()，会让当前goroutine在收到信号前一直处于阻塞状态。
	cond.Wait()
	fmt.Printf("%d\n", s[0])
	cond.L.Unlock()
}
