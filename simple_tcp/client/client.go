package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testbed/logger"
	"time"
)

const (
	addr        = "127.0.0.3"
	port        = 8591
	CLIENT_NUM  = 50
	MSG_SIZE    = 8192
	RECORD_DATA = false
)

func main() {
	/* NF stop signal */
	// go func() {
	// 	time.Sleep(30 * time.Second)
	// 	onvmpoller.CloseONVM()
	// 	os.Exit(1)
	// }()
	// defer onvmpoller.CloseONVM()

	// ID, _ := onvmpoller.IpToID(addr)
	// logger.Log.Infof("[ONVM ID]: %d", ID)

	/* Wait all client finish */
	wg := new(sync.WaitGroup)
	wg.Add(CLIENT_NUM + 3)

	performance := make(chan float64, 10)
	performanceWrite := make(chan float64, 10)
	performanceRead := make(chan float64, 10)
	go calculateRead(wg, performanceWrite)
	go calculateWrite(wg, performanceRead)
	go calculate(wg, performance)

	time.Sleep(5 * time.Second)
	for i := 1; i <= CLIENT_NUM; i++ {
		go client(i, wg, performance, performanceRead, performanceWrite)
		time.Sleep(1 * time.Millisecond)
	}
	wg.Wait()
	time.Sleep(30 * time.Second)
}

func calculate(wg *sync.WaitGroup, performance chan float64) {
	var (
		count  int
		result float64
	)
	defer wg.Done()
	count = 0
	result = 0
	array := make([]string, CLIENT_NUM)

	for {
		select {
		case p := <-performance:
			array[count] = fmt.Sprintf("%.8f", p)
			count++
			result = result + p
		default:
		}

		if count == CLIENT_NUM {
			break
		}
	}

	if RECORD_DATA {
		f, err := os.OpenFile("/home/hstsai/onvm/result/total.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 6666)
		defer f.Close()

		if err != nil {

			logger.Log.Errorln("failed to open file", err)
		}

		w := csv.NewWriter(f)
		defer w.Flush()

		if err := w.Write(array); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	logger.Log.Infof("Average time: %f", result/float64(CLIENT_NUM))
}

func calculateWrite(wg *sync.WaitGroup, performanceWrite chan float64) {
	var (
		count  int
		result float64
	)
	defer wg.Done()
	count = 0
	result = 0
	array := make([]string, CLIENT_NUM)

	for {
		select {
		case p := <-performanceWrite:
			array[count] = fmt.Sprintf("%.8f", p)
			count++
			result = result + p
		default:
		}

		if count == CLIENT_NUM {
			break
		}
	}

	if RECORD_DATA {
		f, err := os.OpenFile("/home/hstsai/onvm/result/write.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 6666)
		defer f.Close()

		if err != nil {

			logger.Log.Errorln("failed to open file", err)
		}

		w := csv.NewWriter(f)
		defer w.Flush()

		if err := w.Write(array); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	logger.Log.Infof("Average read: %f", result/float64(CLIENT_NUM))
}

func calculateRead(wg *sync.WaitGroup, performanceRead chan float64) {
	var (
		count  int
		result float64
	)
	defer wg.Done()
	count = 0
	result = 0
	array := make([]string, CLIENT_NUM)

	for {
		select {
		case p := <-performanceRead:
			array[count] = fmt.Sprintf("%.8f", p)
			count++
			result = result + p
		default:
		}

		if count == CLIENT_NUM {
			break
		}
	}

	if RECORD_DATA {
		f, err := os.OpenFile("/home/hstsai/onvm/result/read.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 6666)
		defer f.Close()

		if err != nil {

			logger.Log.Errorln("failed to open file", err)
		}

		w := csv.NewWriter(f)
		defer w.Flush()

		if err := w.Write(array); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	logger.Log.Infof("Average write: %f", result/float64(CLIENT_NUM))
}

func client(num int, wg *sync.WaitGroup, performance chan float64, performanceRead chan float64, performanceWrite chan float64) {
	defer wg.Done()
	msg := fmt.Sprintf("[Client + Conn_%d]", num)

	t1 := time.Now()
	_, tW, tR, err := sendTCP("127.0.0.2:8591", msg)
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

	conn, err = net.Dial("tcp", addr)
	// conn, err = onvmpoller.DialONVM("onvm", addr)

	if err != nil {
		return "", 0.0, 0.0, err
	}

	defer conn.Close()

	bs := make([]byte, MSG_SIZE)

	BIG_MSG := make([]byte, MSG_SIZE)
	for i := range BIG_MSG {
		BIG_MSG[i] = 87
	}

	t1 := time.Now()
	conn.Write(BIG_MSG)
	t2 := time.Now()
	length, err := conn.Read(bs)
	t3 := time.Now()
	tW := t2.Sub(t1).Seconds() * 1000
	tR := t3.Sub(t2).Seconds() * 1000

	// if string(bs) == string(BIG_MSG) {
	// 	logger.Log.Warnf("=====EQUAL=====")
	// }

	if err != nil {
		return "", tW, tR, err
	} else {
		// return "", tW, tR, err
		return string(bs[:length]), tW, tR, err
	}
}
