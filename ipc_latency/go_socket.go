package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

type Four_tuple_rte struct {
	Src_ip   uint32
	Src_port uint16
	Dst_ip   uint32
	Dst_port uint16
}

type ChannelData struct {
	PacketType  int
	FourTuple   Four_tuple_rte
	Payload_len int
	// PacketList  []*C.struct_rte_mbuf
	PacketList *int
} // 40

const TIMES int = 10

func server() {
	listener, _ := net.Listen("tcp", "127.0.0.1:50010")
	conn, _ := listener.Accept()
	defer conn.Close()

	b := make([]byte, 40)
	for i := 0; i < TIMES; i++ {
		conn.Read(b)
		t := time.Now()
		conn.Write(b)
		fmt.Printf("%v\n", t.Nanosecond())
	}
}

func client() {
	conn, _ := net.Dial("tcp", "127.0.0.1:50010")
	defer conn.Close()

	b := make([]byte, 40)
	for i := 0; i < 40; i++ {
		b[i] = 'X'
	}
	for i := 0; i < TIMES; i++ {
		t1 := time.Now()
		conn.Write(b)
		conn.Read(b)
		t2 := time.Now()
		fmt.Printf("%v - %v = %v (ns)\n", t1.Nanosecond(), t2.Nanosecond(), t2.Sub(t1).Nanoseconds())
		time.Sleep(1 * time.Microsecond)
	}
}

func main() {
	if os.Args[1] == "s" {
		server()
	} else {
		client()
	}
}
