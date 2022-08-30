package main

import (
	"fmt"
	"os"
	"sync"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	addr       = "127.0.0.1"
	port       = 6000
	CLIENT_NUM = 10
)

func main() {
	/* NF stop signal */
	go func() {
		time.Sleep(60 * time.Second)
		onvmpoller.CloseONVM()
		os.Exit(1)
	}()
	defer onvmpoller.CloseONVM()

	ID, _ := onvmpoller.IpToID(addr)
	logger.Log.Infof("[ONVM ID]: %d", ID)

	/* Wait all client finish */
	wg := new(sync.WaitGroup)
	wg.Add(CLIENT_NUM)

	for i := 1; i <= CLIENT_NUM; i++ {
		go client(i, wg)
		time.Sleep(1 * time.Millisecond)
	}
	wg.Wait()
}

func client(num int, wg *sync.WaitGroup) {
	defer wg.Done()
	msg := fmt.Sprintf("This is client%d", num)
	res, err := sendTCP("127.0.0.2:8000", msg)
	if err != nil {
		logger.Log.Errorln(err.Error())
	} else {
		logger.Log.Infof("Recv response: %+v", res)
	}
	time.Sleep(10 * time.Second)
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
