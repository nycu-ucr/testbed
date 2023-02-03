package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

var (
	N_CONNS         = 8
	ADDRESS         = [4]byte{127, 0, 0, 1}
	BASE_PORT       = 59010
	ROUNDTRIP_TIMES = 10000
)

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

func calcSD(num []float64) float64 {
	var sum, mean, sd float64

	mean = sum / float64(ROUNDTRIP_TIMES)

	for j := 0; j < ROUNDTRIP_TIMES; j++ {
		// The use of Pow math function func Pow(x, y float64) float64
		sd += math.Pow(num[j]-mean, 2)
	}
	// The use of Sqrt math function func Sqrt(x float64) float64
	sd = math.Sqrt(sd / float64(ROUNDTRIP_TIMES))

	return sd
}

func server() {
	wg := new(sync.WaitGroup)
	wg.Add(N_CONNS)

	for i := 0; i < N_CONNS; i++ {
		go func(port int) {
			server_address := &unix.SockaddrInet4{
				Port: port,
				Addr: ADDRESS,
			}

			listen_sock, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
			if err != nil {
				log.Fatal("Create Socket: ", err)
				os.Exit(1)
			}

			err = unix.Bind(listen_sock, server_address)
			if err != nil {
				log.Fatal("Bind: ", err)
				os.Exit(1)
			}
			// fmt.Printf("Server: Bind to addr: %d, port: %d\n", server_address.Addr, server_address.Port)

			err = unix.Listen(listen_sock, 1024)
			if err != nil {
				log.Fatal("Listen: ", err)
				os.Exit(1)
			}

			client_sock, _, err := unix.Accept(listen_sock)
			if err != nil {
				log.Fatal("Accept: ", err)
				os.Exit(1)
			}

			buffer := make([]byte, 2048)
			for {
				bytes_read, _ := unix.Read(client_sock, buffer)
				if bytes_read == 0 {
					unix.Close(client_sock)
					break
				} else {
					_, err := unix.Write(client_sock, buffer[:bytes_read])
					if err != nil {
						log.Fatal("Write: ", err)
						os.Exit(1)
					}
				}
			}

			unix.Close(listen_sock)
			wg.Done()
		}(i + BASE_PORT)
	}

	wg.Wait()
}

func client() {
	latency_ch := make(chan float64)

	for i := 0; i < N_CONNS; i++ {
		go func(port int) {
			server_address := &unix.SockaddrInet4{
				Port: port,
				Addr: ADDRESS,
			}

			client_sock, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
			if err != nil {
				log.Fatal("Create Socket: ", err)
				os.Exit(1)
			}

			err = unix.Connect(client_sock, server_address)
			if err != nil {
				log.Fatal("Connect: ", err)
				os.Exit(1)
			}

			var t1, t2 int64
			var latency_sum float64 = 0.0
			latencies := make([]float64, ROUNDTRIP_TIMES)

			for x := 0; x < ROUNDTRIP_TIMES; x++ {
				buffer := makeMsg(64)
				if err != nil {
					fmt.Println("binary.Write failed:", err)
				}

				_, err = unix.Write(client_sock, buffer)
				if err != nil {
					log.Fatal("Client Write: ", err)
				}
				_, err = unix.Read(client_sock, buffer)
				if err != nil {
					log.Fatal("Client Read: ", err)
				}

				t2 = time.Now().UnixNano()
				t1 = int64(parseMsg(buffer))
				latencies[x] = float64(t2-t1) / 1000.0

				latency_sum += float64(t2-t1) / 1000.0
			}
			unix.Close(client_sock)
			// fmt.Printf("SD: %.3f\n", calcSD(latencies))

			latency_ch <- latency_sum / float64(ROUNDTRIP_TIMES)

		}(i + BASE_PORT)
	}

	latency_sum := 0.0
	x := 0
	for v := range latency_ch {
		latency_sum += v

		x++
		if x == N_CONNS {
			break
		}
	}

	fmt.Printf("%.3f\n", latency_sum/float64(N_CONNS))
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: prog [s | c]")
		os.Exit(0)
	}
	if os.Args[1] == "s" {
		server()
	} else {
		client()
	}
}
