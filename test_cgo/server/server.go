package main

import (
	"sync"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

var (
	msg_size int
	result   []int64
)

func client(wg *sync.WaitGroup) {
	for i := 0; i < 100; i++ {
		onvmpoller.TestCgoCondWait(i)
	}
	wg.Done()
}

func main() {

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go client(wg)

	time.Sleep(1 * time.Second)
	for i := 0; i < 100; i++ {
		onvmpoller.TestCgoCondSignal(i)
		time.Sleep(1 * time.Second)
	}
	wg.Wait()

	logger.Log.Infof("Program End")
}
