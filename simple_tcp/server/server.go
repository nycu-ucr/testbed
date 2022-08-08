package main

import (
	"fmt"
	"net"
	"strconv"
)

const (
	addr = ""
	port = 8000
)

func main() {

	src := addr + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", src)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer listener.Close()
	fmt.Printf("TCP server start and listening on %s.\n", src)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Some connection error: %s\n", err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	fmt.Println("Client connected from: " + remoteAddr)

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	for {
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err != nil {

			if err.Error() == "EOF" {
				fmt.Println("Disconned from ", remoteAddr)
				break
			} else {
				fmt.Println("Error reading:", err.Error())
				break
			}
		}
		// Send a response back to person contacting us.
		conn.Write([]byte("Message received.\n"))

		fmt.Printf("len: %d, recv: %s\n", reqLen, string(buf[:reqLen]))
	}
	// Close the connection when you're done with it.
	conn.Close()
}
