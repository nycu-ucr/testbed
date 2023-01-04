package main

import (
	"flag"
	"runtime"
	"strconv"
	"sync"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

var (
	server_addr = "127.0.0.2"
	server_port = 8591
	msg_size    int
	loop_times  int
	result      []int64
)

func main() {
	runtime.GOMAXPROCS(2)
	flag.IntVar(&msg_size, "m", 64, "Setup Message Size (Default is 64)")
	flag.IntVar(&loop_times, "lt", 5, "Setup Loop Times (Default is 5)")
	flag.Parse()
	logger.Log.Warnf("[MSG_Size: %d][LOOP_NUM: %d]", msg_size, loop_times)
	wg := &sync.WaitGroup{}
	wg.Add(loop_times)
	result = make([]int64, loop_times)

	server := server_addr + ":" + strconv.Itoa(server_port)

	for i := 0; i < loop_times; i++ {
		go client(i, server, wg)
	}

	wg.Wait()
	logger.Log.Infof("Program End")
	total := 0
	for i := 0; i < loop_times; i++ {
		total = total + int(result[i])
	}
	logger.Log.Warnf("Total roundtrip count: %d/s", total)
	time.Sleep(10 * time.Second)

	// onvmpoller.CloseONVM()
}

func client(client_ID int, server string, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := onvmpoller.DialONVM("onvm", server)
	// conn, err := net.Dial("tcp", server)
	if err != nil {
		println(err.Error())
	}

	start := time.Now()
	interval := start.Add(1 * time.Second)
	a_size := 0
	loopNum := 0

	for {
		n, err := conn.Write(makeMsg(msg_size))
		a_size += n
		if err != nil {
			logger.Log.Errorf("Write error: %+v", err)
			break
		}

		buf := make([]byte, msg_size)
		_, err = conn.Read(buf)
		if err != nil {
			logger.Log.Errorf("Read error: %+v", err)
			break
		}

		loopNum++

		if time.Now().After(interval) {
			break
		}
	}

	conn.Close()

	result[client_ID] = int64(loopNum)
	logger.Log.Infof("[Client %d]Roundtrip count: %d/s", client_ID, loopNum)
}

func makeMsg(msg_size int) []byte {
	b := make([]byte, msg_size)
	v := uint64(time.Now().Nanosecond())

	for i := 0; i < 8; i++ {
		b[0] = byte(v >> 56)
		b[1] = byte(v >> 48)
		b[2] = byte(v >> 40)
		b[3] = byte(v >> 32)
		b[4] = byte(v >> 24)
		b[5] = byte(v >> 16)
		b[6] = byte(v >> 8)
		b[7] = byte(v)
	}
	for j := 8; j < msg_size; j++ {
		b[j] = 87
	}

	return b
}
