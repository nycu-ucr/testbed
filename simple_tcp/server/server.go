package main

import (
	"fmt"
	"net"
	"strconv"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	addr = "127.0.0.11"
	port = 8000
)

func main() {
	/* NF stop signal */

	server_chan := make(chan bool)
	client_chan := make(chan bool)

	go server(server_chan)
	go client(client_chan)

	<-client_chan
	<-server_chan

	logger.Log.Infoln("After 10 second, shutdown the process")
	time.Sleep(10 * time.Second)
	onvmpoller.CloseONVM()
}

func server(server_chan chan bool) {
	src := addr + ":" + strconv.Itoa(port)
	ID, _ := onvmpoller.IpToID(addr)
	logger.Log.Infof("[ONVM ID]: %d", ID)

	listener, err := onvmpoller.ListenONVM("onvm", src)
	if err != nil {
		logger.Log.Errorln(err.Error())
	}
	defer listener.Close()
	logger.Log.Infof("TCP server start and listening on %s", src)

	// Handle single connection
	conn, err := listener.Accept()
	if err != nil {
		logger.Log.Errorf("Some connection error: %s\n", err)
	}

	handleConnection(conn)

	server_chan <- true
}

func handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	logger.Log.Infof("Client connected from: " + remoteAddr)

	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	for {
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err != nil {

			if err.Error() == "EOF" {
				logger.Log.Infof("Disconned from: %s", remoteAddr)
				break
			} else {
				logger.Log.Errorf("Error reading:", err.Error())
				break
			}
		} else {
			// Send a response back to person contacting us.
			msg := fmt.Sprintf("Message received: %s\n", string(buf[:reqLen]))
			conn.Write([]byte(msg))

			logger.Log.Infof("len: %d, recv: %s\n", reqLen, string(buf[:reqLen]))
		}
	}
	logger.Log.Infof("Client close connection")
	// Close the connection when you're done with it.
	conn.Close()
}

func client(client_chan chan bool) {
	time.Sleep(5 * time.Second)
	conn, err := onvmpoller.DialONVM("onvm", "127.0.0.1:8000")
	msg_count := 5
	msg := ""
	buf := make([]byte, 256)
	prefix := "\u001b[33m[I'm server]\u001b[0m"

	if err != nil {
		logger.Log.Errorln(err.Error())
	}

	for i := 0; i < msg_count; i++ {
		msg = fmt.Sprintf("%s Message %v", prefix, i+1)
		conn.Write([]byte(msg))

		_, err := conn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				logger.Log.Infoln("Peer close connection")
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	conn.Close()
	client_chan <- true
}
