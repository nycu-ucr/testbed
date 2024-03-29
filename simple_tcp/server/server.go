package main

import (
	"net"
	"strconv"
	"testbed/logger"
)

const (
	addr     = "127.0.0.2"
	port     = 8591
	MSG_SIZE = 8192
)

func main() {
	var listener net.Listener
	var err error
	src := addr + ":" + strconv.Itoa(port)

	/* NF stop signal */
	// go func() {
	// 	time.Sleep(30 * time.Second)
	// 	onvmpoller.CloseONVM()
	// 	os.Exit(1)
	// }()
	// defer onvmpoller.CloseONVM()
	// ID, _ := onvmpoller.IpToID(addr)
	// logger.Log.Infof("[ONVM ID]: %d", ID)
	// listener, err = onvmpoller.ListenONVM("onvm", src)

	listener, err = net.Listen("tcp", src)

	if err != nil {
		logger.Log.Errorln(err.Error())
	}
	defer listener.Close()
	logger.Log.Infof("TCP server start and listening on %s", src)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Log.Errorf("Some connection error: %s\n", err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// remoteAddr := conn.RemoteAddr().String()
	// logger.Log.Infof("Client connected from: " + remoteAddr)

	// Make a buffer to hold incoming data.
	buf := make([]byte, MSG_SIZE)
	for {
		// Read the incoming connection into the buffer.
		reqLen, err := conn.Read(buf)
		if err != nil {

			if err.Error() == "EOF" {
				// logger.Log.Infof("Disconned from: %s", remoteAddr)
				break
			} else {
				logger.Log.Errorf("Error reading:", err.Error())
				break
			}
		} else {
			// Send a response back to person contacting us.
			msg := string(buf[:reqLen])
			conn.Write([]byte(msg))

			// logger.Log.Infof("len: %d, recv:\n%s\n", reqLen, string(buf[:reqLen]))
			// logger.Log.Infof("len: %d\n", reqLen)
		}
	}
	// logger.Log.Infof("Client close connection")
	// Close the connection when you're done with it.
	conn.Close()
}
