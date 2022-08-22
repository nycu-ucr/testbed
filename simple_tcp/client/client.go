package main

import (
	"os"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	addr = "127.0.0.1"
	port = 6000
)

func main() {
	ID, _ := onvmpoller.IpToID(addr)
	logger.Log.Infof("[ONVM ID]: %d", ID)

	res, err := sendTCP("127.0.0.2:8000", "HA HA is me")
	if err != nil {
		logger.Log.Errorln(err.Error())
	} else {
		logger.Log.Infof("Recv response: %+v", res)
	}

	time.Sleep(30 * time.Second)
	onvmpoller.CloseONVM()
	os.Exit(1)
}

func sendTCP(addr, msg string) (string, error) {
	// connect to this socket
	conn, err := onvmpoller.DialONVM("onvm", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// send to socket
	conn.Write([]byte(msg))

	// listen for reply
	bs := make([]byte, 1024)
	len, err := conn.Read(bs)
	if err != nil {
		return "", err
	} else {
		return string(bs[:len]), err
	}
}
