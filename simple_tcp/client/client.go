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
	CLIENT_NUM = 100
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

	/* Wait all client finish */
	wg_conn, wg_worker := new(sync.WaitGroup), new(sync.WaitGroup)
	wg_conn.Add(CLIENT_NUM)
	wg_worker.Add(5)

	performance := make(chan float64, 10)
	performanceWrite := make(chan float64, 10)
	performanceRead := make(chan float64, 10)
	performanceCreate := make(chan float64, 10)
	performanceClose := make(chan float64, 10)

	go calculate(wg_worker, performance)
	go calculateCreate(wg_worker, performanceCreate, performance)
	go calculateRead(wg_worker, performanceRead, performance)
	go calculateWrite(wg_worker, performanceWrite, performance)
	go calculateClose(wg_worker, performanceClose, performance)

	conns := [CLIENT_NUM + 1]net.Conn{}

	time.Sleep(3 * time.Second)
	for i := 1; i <= CLIENT_NUM; i++ {
		go create_connection(wg_conn, performanceCreate, &conns, i)
	}
	wg_conn.Wait()
	wg_conn.Add(CLIENT_NUM)

	time.Sleep(1 * time.Second)
	for i := 1; i <= CLIENT_NUM; i++ {
		go write_and_read(wg_conn, performanceRead, performanceWrite, i, conns[i])
	}
	wg_conn.Wait()
	wg_conn.Add(CLIENT_NUM)

	time.Sleep(1 * time.Second)
	for i := 1; i <= CLIENT_NUM; i++ {
		go close_connection(wg_conn, performanceClose, conns[i])
	}
	wg_conn.Wait()

	wg_worker.Wait()
	time.Sleep(30 * time.Second)
}

func create_connection(wg *sync.WaitGroup, performanceCreate chan float64, conns *[CLIENT_NUM + 1]net.Conn, num int) {
	defer wg.Done()

	t1 := time.Now()
	conn, err := onvmpoller.DialONVM("onvm", "127.0.0.2:8000")
	// conn, err := net.Dial("tcp", "127.0.0.2:8000")
	t2 := time.Now()

	t := t2.Sub(t1).Seconds() * 1000

	performanceCreate <- t

	conns[num] = conn

	if err != nil {
		fmt.Println(err.Error())
	}
}

func write_and_read(wg *sync.WaitGroup, performanceRead chan float64, performanceWrite chan float64, num int, conn net.Conn) {
	defer wg.Done()

	buf := make([]byte, 1024)
	msg := fmt.Sprintf("[Client + Conn_%d]", num)

	t1 := time.Now()
	conn.Write([]byte(msg))
	t2 := time.Now()
	_, err := conn.Read(buf)
	t3 := time.Now()
	tW := t2.Sub(t1).Seconds() * 1000
	tR := t3.Sub(t2).Seconds() * 1000

	performanceRead <- tR
	performanceWrite <- tW

	if err != nil {
		fmt.Println(err.Error())
	}
}

func close_connection(wg *sync.WaitGroup, performanceClose chan float64, conn net.Conn) {
	defer wg.Done()

	t1 := time.Now()
	conn.Close()
	t2 := time.Now()

	t := t2.Sub(t1).Seconds() * 1000

	performanceClose <- t
}

func calculateCreate(wg *sync.WaitGroup, performanceCreate chan float64, performance chan float64) {
	var (
		count  int
		result float64
	)
	defer wg.Done()
	count = 0
	result = 0

	for {
		select {
		case p := <-performanceCreate:
			count++
			result = result + p
		default:
		}

		if count == CLIENT_NUM {
			break
		}
	}

	v := result / float64(CLIENT_NUM)
	logger.Log.Infof("Average create connection: %f", v)
	performance <- v
}

func calculateWrite(wg *sync.WaitGroup, performanceWrite chan float64, performance chan float64) {
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

	v := result / float64(CLIENT_NUM)
	logger.Log.Infof("Average write: %f", v)
	performance <- v
}

func calculateRead(wg *sync.WaitGroup, performanceRead chan float64, performance chan float64) {
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

	v := result / float64(CLIENT_NUM)
	logger.Log.Infof("Average read: %f", v)
	performance <- v
}

func calculateClose(wg *sync.WaitGroup, performanceClose chan float64, performance chan float64) {
	var (
		count  int
		result float64
	)
	defer wg.Done()
	count = 0
	result = 0

	for {
		select {
		case p := <-performanceClose:
			count++
			result = result + p
		default:
		}

		if count == CLIENT_NUM {
			break
		}
	}

	v := result / float64(CLIENT_NUM)
	logger.Log.Infof("Average close connection: %f", v)
	performance <- v
}

func calculate(wg *sync.WaitGroup, performance chan float64) {
	defer wg.Done()

	count := 0
	worker := 4
	result := 0.0

	for {
		select {
		case p := <-performance:
			count++
			result = result + p
		default:
		}

		if count == worker {
			break
		}
	}

	logger.Log.Infof("Average time: %f", result)
}
