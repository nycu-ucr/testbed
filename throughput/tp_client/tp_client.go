package main

import (
	"encoding/binary"
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

const (
	CLIENT_LOOP_TIMES = 1000
	unix_socket_addr  = "test.sock"
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

	// t := time.Now()
	for i := 0; i < loop_times; i++ {
		go client(i, server, wg)
	}

	wg.Wait()
	// total_latency := time.Since(t).Nanoseconds()
	logger.Log.Infof("Program End")
	total := int64(0)
	for i := 0; i < loop_times; i++ {
		total = total + result[i]
	}
	logger.Log.Warnf("Average roundtrip latency: %d(ns)", total/int64(loop_times))
	// logger.Log.Warnf("Total time: %d(ns)", total_latency)
	// time.Sleep(10 * time.Second)

	// onvmpoller.CloseONVM()
}

func client(client_ID int, server string, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := onvmpoller.DialONVM("onvm", server)
	// conn, err := net.Dial("tcp", server)
	// conn, err := net.Dial("unix", unix_socket_addr)
	if err != nil {
		println(err.Error())
	}
	// time.Sleep(1 * time.Second)

	// arrival_distribution := distuv.Poisson{
	// 	Lambda: 2.0,
	// 	Src:    rand.NewSource(uint64(time.Now().UnixNano())),
	// }

	// start := time.Now()
	// interval := start.Add(1 * time.Second)
	roundtrip := int64(0)
	a_size := 0
	loopNum := 0

	for i := 0; i < CLIENT_LOOP_TIMES; i++ {
		n, err := conn.Write(makeMsg(msg_size))
		a_size += n
		if err != nil {
			logger.Log.Errorf("Write error: %+v", err)
			break
		}

		buf := make([]byte, msg_size)
		_, err = conn.Read(buf)
		t2 := time.Now().UnixNano()
		t1 := parseMsg(buf)
		if err != nil {
			logger.Log.Errorf("Read error: %+v", err)
			break
		}

		loopNum++

		// if time.Now().After(interval) {
		// 	break
		// }
		roundtrip = roundtrip + (int64(t2) - int64(t1))
		// logger.Log.Infof("delay: %d", (int64(t2) - int64(t1)))
		// time.Sleep((time.Duration(arrival_distribution.Rand()) + 20) * time.Microsecond)
		// time.Sleep(10 * time.Millisecond)
		// ts := time.Now()
		// for a := 0; a < 100000; a++ {

		// }
		// te := time.Since(ts).Nanoseconds()
		// logger.Log.Infof("for loop time: %d(ns)", te)
	}

	conn.Close()

	result[client_ID] = roundtrip / int64(loopNum)
	logger.Log.Infof("Loop num: %d", loopNum)
	logger.Log.Infof("[Client %d]Roundtrip latency: %d(ns)", client_ID, roundtrip/int64(loopNum))
}

func makeMsg(msg_size int) []byte {
	b := make([]byte, msg_size)

	for j := 8; j < msg_size; j++ {
		b[j] = 87
	}

	v := uint64(time.Now().UnixNano())
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

	return b
}

func parseMsg(b []byte) uint64 {
	return binary.BigEndian.Uint64(b[0:8])
}
