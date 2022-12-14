package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	server_addr = "127.0.0.2"
	server_port = 8591
)

func main() {
	var msg_size int
	var loop_times int

	flag.IntVar(&msg_size, "m", 64, "Setup Message Size (Default is 64)")
	flag.IntVar(&loop_times, "lt", 5, "Setup Loop Times (Default is 5)")
	flag.Parse()

	server := server_addr + ":" + strconv.Itoa(server_port)

	buf := make([]byte, msg_size)
	for x := range buf {
		buf[x] = 1
	}

	for i := 0; i < loop_times; i++ {
		// conn, err := onvmpoller.DialONVM("onvm", server)
		conn, err := net.Dial("tcp", server)
		if err != nil {
			println(err.Error())
			break
		}
		start := time.Now()
		interval := start.Add(2 * time.Second)
		a_size := 0

		for {
			n, err := conn.Write(buf)
			a_size += n
			if err != nil {
				break
			}

			if time.Now().After(interval) {
				break
			}
		}
		end := time.Now()

		duration := end.Sub(start).Seconds()

		mbps := (float64(a_size) / (duration * 1000000))

		conn.Close()
		fmt.Printf("Loop:%d\tMessage Size: %d\tMbps: %.3f\n", i+1, msg_size, mbps)
	}

	println("Program End")
	time.Sleep(10 * time.Second)

	// onvmpoller.CloseONVM()
}
