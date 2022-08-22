package main

import (
	"net"
	"os"
	"strconv"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	addr = "127.0.0.2"
	port = 8000
)

func main() {
	src := addr + ":" + strconv.Itoa(port)
	ID, _ := onvmpoller.IpToID(addr)
	logger.Log.Infof("[ONVM ID]: %d", ID)

	listener, err := onvmpoller.ListenONVM("onvm", src)
	if err != nil {
		logger.Log.Errorln(err.Error())
	}
	defer listener.Close()
	logger.Log.Infof("TCP server start and listening on %s", src)

	/* NF stop signal */
	go func() {
		time.Sleep(30 * time.Second)
		onvmpoller.CloseONVM()
		os.Exit(1)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Log.Errorf("Some connection error: %s\n", err)
		}

		go handleConnection(conn)
	}
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
				logger.Log.Infof("Disconned from ", remoteAddr)
				break
			} else {
				logger.Log.Errorf("Error reading:", err.Error())
				break
			}
		}
		// Send a response back to person contacting us.
		conn.Write([]byte("Message received.\n"))

		logger.Log.Infof("len: %d, recv: %s\n", reqLen, string(buf[:reqLen]))
	}
	// Close the connection when you're done with it.
	conn.Close()
}
