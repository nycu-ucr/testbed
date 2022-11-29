package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	addr       = "127.0.0.1"
	port       = 6000
	CLIENT_NUM = 500
)

var (
	read_time  float64
	write_time float64
	total_time float64
)

func main() {
	/* NF stop signal */
	go func() {
		time.Sleep(30 * time.Second)
		onvmpoller.CloseONVM()
		os.Exit(1)
	}()
	defer onvmpoller.CloseONVM()

	ID, _ := onvmpoller.IpToID(addr)
	logger.Log.Infof("[ONVM ID]: %d", ID)
	onvmpoller.SetLocalAddress(addr)

	is_onvm := true
	var fname string
	// Relative to onvm/testbed/bin
	if is_onvm {
		fname = "../simple_tcp/simple_tcp3_onvm.csv"
	} else {
		fname = "../simple_tcp/simple_tcp3_tcp.csv"
	}

	/* Wait all client finish */
	wg := new(sync.WaitGroup)
	wg.Add(CLIENT_NUM + 3)

	performance := make(chan float64, 10)
	performanceWrite := make(chan float64, 10)
	performanceRead := make(chan float64, 10)
	go calculateRead(wg, performanceWrite)
	go calculateWrite(wg, performanceRead)
	go calculate(wg, performance)

	for i := 1; i <= CLIENT_NUM; i++ {
		go client(i, wg, performance, performanceRead, performanceWrite)
		time.Sleep(1 * time.Millisecond)
	}
	wg.Wait()
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	text := fmt.Sprintf("%.6f,%.6f,%.6f\n", read_time, write_time, total_time)
	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}
	// time.Sleep(30 * time.Second)
}

func calculate(wg *sync.WaitGroup, performance chan float64) {
	var (
		count  int
		result float64
	)
	defer wg.Done()
	count = 0
	result = 0

	for {
		select {
		case p := <-performance:
			count++
			result = result + p
		default:
		}

		if count == CLIENT_NUM {
			break
		}
	}

	total_time = result / float64(CLIENT_NUM)
	logger.Log.Infof("Average time: %f (ms)", total_time)
}

func calculateWrite(wg *sync.WaitGroup, performanceWrite chan float64) {
	var (
		count  int
		result float64
	)
	defer wg.Done()
	count = 0
	result = 0

	for {
		select {
		case p := <-performanceWrite:
			count++
			result = result + p
		default:
		}

		if count == CLIENT_NUM {
			break
		}
	}

	read_time = result / float64(CLIENT_NUM)
	logger.Log.Infof("Average read: %f (ms)", read_time)
}

func calculateRead(wg *sync.WaitGroup, performanceRead chan float64) {
	var (
		count  int
		result float64
	)
	defer wg.Done()
	count = 0
	result = 0

	for {
		select {
		case p := <-performanceRead:
			count++
			result = result + p
		default:
		}

		if count == CLIENT_NUM {
			break
		}
	}

	write_time = result / float64(CLIENT_NUM)
	logger.Log.Infof("Average write: %f (ms)", write_time)
}

func client(num int, wg *sync.WaitGroup, performance chan float64, performanceRead chan float64, performanceWrite chan float64) {
	defer wg.Done()
	msg := fmt.Sprintf("[Client + Conn_%d]", num)

	t1 := time.Now()
	_, tW, tR, err := sendTCP("127.0.0.2:8000", msg)
	t2 := time.Now()

	t := t2.Sub(t1).Seconds() * 1000
	performanceRead <- tR
	performanceWrite <- tW
	performance <- t
	// logger.Log.Infof("[Delay] Write=%f, Read=%f", tW, tR)

	if err != nil {
		logger.Log.Errorln(err.Error())
	} else {
		// logger.Log.Infof("[Conn_%d] Recv response: %+v", num, res)
	}
}

func sendTCP(addr, msg string) (string, float64, float64, error) {
	// connect to this socket
	var conn net.Conn
	var err error

	// conn, err = net.Dial("tcp", addr)
	conn, err = onvmpoller.DialONVM("onvm", addr)

	if err != nil {
		return "", 0.0, 0.0, err
	}

	defer conn.Close()

	bs := make([]byte, 1024)

	t1 := time.Now()
	conn.Write([]byte(msg))
	t2 := time.Now()
	length, err := conn.Read(bs)
	t3 := time.Now()
	tW := t2.Sub(t1).Seconds() * 1000
	tR := t3.Sub(t2).Seconds() * 1000

	if err != nil {
		return "", tW, tR, err
	} else {
		// return "", tW, tR, err
		return string(bs[:length]), tW, tR, err
	}
}
