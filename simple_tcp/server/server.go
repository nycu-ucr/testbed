package main

import (
	"flag"
	"net"
	"runtime"
	"strconv"
	"sync"
	"testbed/logger"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	addr             = "127.0.0.2"
	port             = 8591
	READ_WRITE       = true
	THREAD_NUM       = 1
	unix_socket_addr = "test.sock"
)

var (
	msg_size int
	result   []int64
)

func handle_client(client_ID int, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	// eop := errors.New("EOP").Error()
	buf := make([]byte, msg_size)
	_, err := conn.Read(buf)
	// if err != nil && err.Error() != eop {
	// 	logger.Log.Errorf("Read Error: %+v\n", err)
	// }

	_, err = conn.Write(buf)
	if err != nil {
		logger.Log.Errorf("Read error: %+v", err)
	}

	conn.Close()
}

func main() {
	runtime.GOMAXPROCS(2)
	var loop_times int
	wg := &sync.WaitGroup{}

	flag.IntVar(&msg_size, "m", 64, "Setup Message Size (Default is 64)")
	flag.IntVar(&loop_times, "lt", 1, "Setup Loop Times (Default is 1)")
	flag.Parse()
	logger.Log.Warnf("[MSG_Size: %d][LOOP_NUM: %d]", msg_size, loop_times)
	wg.Add(loop_times * THREAD_NUM)
	result = make([]int64, loop_times)

	src := addr + ":" + strconv.Itoa(port)

	listener, _ := onvmpoller.ListenXIO("onvm", src)
	// listener, _ := net.Listen("tcp", src)
	// listener, _ := net.Listen("unix", unix_socket_addr)

	for i := 0; i < loop_times*THREAD_NUM; i++ {
		conn, _ := listener.Accept()
		go handle_client(i, conn, wg)
	}

	wg.Wait()
	logger.Log.Infof("Program End")
	// time.Sleep(10 * time.Second)
}
