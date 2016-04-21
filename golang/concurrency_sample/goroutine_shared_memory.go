package main

import (
	"fmt"
	"sync"
	"runtime"
)

var counter int = 0

func Count(lock *sync.Mutex) {
	lock.Lock()
	counter++
	fmt.Println(counter)
	lock.Unlock()
}

func main() {
	lock := &sync.Mutex{}

	for i := 0; i < 10; i++ {
		go Count(lock)
	}
    //等待所有count运算结束
	for {
		lock.Lock()
		c := counter
		lock.Unlock()

		runtime.Gosched() //出让时间片
		if c >= 10 {
			break
		}
	}
}