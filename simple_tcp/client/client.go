package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

var (
	server_addr        = "127.0.0.2"
	server_port        = 8591
	msg_size           int
	loop_times         int
	result_roundtrip   []int64
	result_conn_create []int64
	result_conn_close  []int64
)

const (
	THREAD_NUM       = 1
	unix_socket_addr = "test.sock"
)

func main() {
	runtime.GOMAXPROCS(2)

	flag.IntVar(&msg_size, "m", 64, "Setup Message Size (Default is 64)")
	flag.IntVar(&loop_times, "lt", 1, "Setup Loop Times (Default is 1)")
	flag.Parse()
	logger.Log.Warnf("[MSG_Size: %d][LOOP_NUM: %d]", msg_size, loop_times)
	wg := &sync.WaitGroup{}
	wg.Add(loop_times * THREAD_NUM)

	result_conn_create = make([]int64, loop_times*THREAD_NUM)
	result_roundtrip = make([]int64, loop_times*THREAD_NUM)
	result_conn_close = make([]int64, loop_times*THREAD_NUM)

	server := server_addr + ":" + strconv.Itoa(server_port)

	for t := 0; t < THREAD_NUM; t++ {
		start := loop_times * t
		end := loop_times * (t + 1)
		go func() {
			for i := start; i < end; i++ {
				client(i, server, wg)
			}
		}()
	}

	wg.Wait()
	logger.Log.Infof("Program End")

	average_create := int64(0)
	average_roundtrip := int64(0)
	average_close := int64(0)

	for i := 0; i < loop_times*THREAD_NUM; i++ {
		average_create = average_create + result_conn_create[i]
		average_roundtrip = average_roundtrip + result_roundtrip[i]
		average_close = average_close + result_conn_close[i]

	}

	create_latency := int(average_create) / (loop_times * THREAD_NUM)       // ns
	roundtrip_latency := int(average_roundtrip) / (loop_times * THREAD_NUM) // ns
	close_latency := int(average_close) / (loop_times * THREAD_NUM)         // ns

	logger.Log.Warnf("Average create latency: %d(ns)", create_latency)
	logger.Log.Warnf("Average roundtrip latency: %d(ns)", roundtrip_latency)
	logger.Log.Warnf("Average close latency: %d(ns)", close_latency)
	// MBs := float32(math.Pow(10, 9)) / float32(l) * float32(msg_size) / float32(math.Pow(10, 6))
	// logger.Log.Warnf("Average throughput: %f MB/s", MBs*2)
	// time.Sleep(10 * time.Second)
	file, err := os.OpenFile("/home/hstsai/onvm/testbed/analyze/short.csv", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	writer := csv.NewWriter(file)
	data := []int64{int64(create_latency), int64(roundtrip_latency), int64(close_latency)}
	var s []string
	for _, value := range data {
		s = append(s, strconv.FormatInt(value, 10))
	}
	err = writer.Write(s)
	if err != nil {
		log.Fatal(err)
	}
	writer.Flush()
	file.Close()

	time.Sleep(1 * time.Second)
}

func client(client_ID int, server string, wg *sync.WaitGroup) {
	defer wg.Done()
	// eop := errors.New("EOP").Error()
	t1_create := time.Now()
	conn, err := onvmpoller.DialXIO("onvm", server)
	// conn, err := net.Dial("tcp", server)
	// conn, err := net.Dial("unix", unix_socket_addr)
	t2_create := time.Since(t1_create)

	if err != nil {
		println(err.Error())
	}

	_, err = conn.Write(makeMsg(msg_size))
	if err != nil {
		logger.Log.Errorf("Write error: %+v", err)
	}

	buf := make([]byte, msg_size)
	_, err = conn.Read(buf)
	// if err != nil && err.Error() != eop {
	// 	logger.Log.Errorf("Read error: %+v", err)
	// }

	t2_r := time.Now().UnixNano()
	t1_r := parseMsg(buf)

	t1_close := time.Now()
	conn.Close()
	t2_close := time.Since(t1_close)

	result_conn_create[client_ID] = t2_create.Nanoseconds()
	result_roundtrip[client_ID] = t2_r - int64(t1_r)
	result_conn_close[client_ID] = t2_close.Nanoseconds()
	// logger.Log.Infof("[Client %d] Average Roundtrip Latency: %d(ns)", client_ID, result[client_ID])
}

func makeMsg(msg_size int) []byte {
	b := make([]byte, msg_size)
	v := uint64(time.Now().UnixNano())

	for i := 0; i < 8; i++ {
		b[0] = byte(v >> 56)
		b[1] = byte(v >> 48)
		b[2] = byte(v >> 40)
		b[3] = byte(v >> 32)
		b[4] = byte(v >> 24)
		b[5] = byte(v >> 16)
		b[6] = byte(v >> 8)
		b[7] = byte(v)
	}
	for j := 8; j < msg_size; j++ {
		b[j] = 87
	}

	return b
}

func parseMsg(b []byte) uint64 {
	return binary.BigEndian.Uint64(b[0:8])
}
