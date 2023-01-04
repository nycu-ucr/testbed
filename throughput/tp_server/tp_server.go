package main

import (
	"flag"
	"net"
	"runtime"
	"strconv"
	"sync"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	addr = "127.0.0.2"
	port = 8591
)

var (
	msg_size int
	result   []int64
)

func handle_client(client_ID int, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	a_size := 0
	start := time.Now()
	loopNum := 0

	for {
		buf := make([]byte, msg_size)
		n, err := conn.Read(buf)
		a_size += n
		if err != nil {
			logger.Log.Errorf("Read Error: %+v\n", err)
			break
		}

		_, err = conn.Write(buf)
		if err != nil {
			logger.Log.Errorf("Read error: %+v", err)
			break
		}

		loopNum++
	}

	duration := time.Since(start)
	MBs := (float64(a_size) / (duration.Seconds() * 1000000))

	logger.Log.Warnf("[CLIENT_ID: %d]", client_ID)
	logger.Log.Infof("MB/s: %.3f", MBs)
	logger.Log.Infof("Loop num: %d", loopNum)
	logger.Log.Infof("Total read bytes: %d", a_size)
	logger.Log.Infof("Duration: %d(ns)", duration.Nanoseconds())
	logger.Log.Infof("Latency: %d(ns)", duration.Nanoseconds()/int64(loopNum))

	conn.Close()
}

func main() {
	runtime.GOMAXPROCS(2)
	var loop_times int
	wg := &sync.WaitGroup{}

	flag.IntVar(&msg_size, "m", 64, "Setup Message Size (Default is 64)")
	flag.IntVar(&loop_times, "lt", 5, "Setup Loop Times (Default is 5)")
	flag.Parse()
	logger.Log.Warnf("[MSG_Size: %d][LOOP_NUM: %d]", msg_size, loop_times)
	wg.Add(loop_times)

	src := addr + ":" + strconv.Itoa(port)
	listener, _ := onvmpoller.ListenONVM("onvm", src)
	// listener, _ := net.Listen("tcp", src)

	for i := 0; i < loop_times; i++ {
		conn, _ := listener.Accept()
		go handle_client(i, conn, wg)
	}

	wg.Wait()
	logger.Log.Infof("Program End")
	time.Sleep(10 * time.Second)
	// onvmpoller.CloseONVM()
}
