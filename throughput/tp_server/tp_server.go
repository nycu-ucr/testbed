package main

import (
	"flag"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	addr        = "127.0.0.2"
	port        = 8591
	BUFFER_SIZE = 4096
)

func handle_client(conn net.Conn, wg *sync.WaitGroup) {
	buf := make([]byte, BUFFER_SIZE)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
	}
	conn.Close()
	wg.Done()
}

func main() {
	var loop_times int
	wg := &sync.WaitGroup{}

	flag.IntVar(&loop_times, "lt", 5, "Setup Loop Times (Default is 5)")
	flag.Parse()
	wg.Add(loop_times)

	src := addr + ":" + strconv.Itoa(port)
	// listener, _ := onvmpoller.ListenONVM("onvm", src)
	listener, _ := net.Listen("tcp", src)

	for i := 0; i < loop_times; i++ {
		conn, _ := listener.Accept()
		go handle_client(conn, wg)
	}

	wg.Wait()
	println("Program End")
	time.Sleep(10 * time.Second)
	// onvmpoller.CloseONVM()
}
