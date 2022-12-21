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
	var packet_count int

	flag.IntVar(&msg_size, "m", 64, "Setup Message Size (Default is 64)")
	flag.IntVar(&loop_times, "lt", 1, "Setup Loop Times (Default is 5)")
	flag.IntVar(&packet_count, "pc", 10000, "Setup Packet Counts (Default is 10000)")
	flag.Parse()

	server := server_addr + ":" + strconv.Itoa(server_port)

	buf := make([]byte, msg_size)
	rbuf := make([]byte, 3)
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

		a_size := 0
		start := time.Now()
		for i := 0; i < packet_count; i++ {
			n, err := conn.Write(buf)
			a_size += n
			if err != nil {
				fmt.Printf("Write: %v", err.Error())
				break
			}
			_, err = conn.Read(rbuf)
			if err != nil {
				fmt.Printf("Read: %v", err.Error())
			}
			// fmt.Printf("\rPacket Count: %d", i+1)
		}
		end := time.Now()
		fmt.Println()

		duration := end.Sub(start).Seconds()
		mbps := (float64(a_size) / (duration * 1000000))

		conn.Close()
		fmt.Printf("Loop:%d\tMessage Size: %d\tMbps: %.3f\n", i+1, msg_size, mbps)
	}

	println("Program End")
	time.Sleep(10 * time.Second)

	// onvmpoller.CloseONVM()
}
