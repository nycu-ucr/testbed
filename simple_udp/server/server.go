package main

import (
	"fmt"
	"net"
	"os"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	server_addr = "127.0.0.2"
	server_port = 8000
)

func get_udp_conn() *onvmpoller.UDP_Connection {
	udp_addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", server_addr, server_port))
	if err != nil {
		fmt.Printf("ResolveUDPAddr, error: %v\n", err)
		os.Exit(1)
	}

	udp_conn, err := onvmpoller.ListenXIO_UDP("onvm-udp", udp_addr)
	if err != nil {
		fmt.Printf("ListenXIO_UDP, error: %v\n", err)
		os.Exit(1)
	}

	return udp_conn
}

func main() {
	udp_conn := get_udp_conn()

	buf := make([]byte, 2048)
	n, remote, err := udp_conn.ReadFrom(buf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Printf("From %s, read %d bytes\n\tMsg: %s\n", remote.String(), n, string(buf))

	udp_conn.WriteTo([]byte("over"), remote)
}
