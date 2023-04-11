package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/gonet/http"
	"github.com/nycu-ucr/net/http2"
	"github.com/nycu-ucr/onvmpoller"
	"github.com/sirupsen/logrus"
)

const (
	EPOCHS             = 1
	USE_ONVM           = false
	USE_ONVM_TRANSPORT = false
)

var (
	ID       int
	datas    []string
	n_thread int
	n_req    int
)

type User struct {
	Id       string `json:"UserId"`
	Name     string `json:"UserName"`
	Password string `json:"UserPassword"`
}

func main() {
	// if USE_ONVM {
	// 	go func() {
	// 		time.Sleep(30 * time.Second)
	// 		onvmpoller.CloseONVM()
	// 		os.Exit(1)
	// 	}()
	// 	defer onvmpoller.CloseONVM()
	// }
	onvmpoller.SetLocalAddress("127.0.0.3")

	/* Init global var */
	logger.Log.Logger.SetLevel(logrus.InfoLevel)
	ID = 50
	datas = make([]string, 0)

	flag.IntVar(&n_thread, "t", 1, "Setup # of thread (Default is 1)")
	flag.IntVar(&n_req, "req", 100000, "Setup # of request per thread (Default is 100000)")
	flag.Parse()

	for i := 0; i < EPOCHS; i++ {
		// http2Post("http://127.0.0.2:8000/test-server/PostUser")
		// http2GET("http://127.0.0.2:8000/test-server/GetUser")
		// http2Post("http://127.0.0.2:8000/test-server/PostUser")
		// http2GET("http://127.0.0.2:8000/test-server/GetUser")
		// http2GET("http://127.0.0.2:8000/test-server/GetUser/1")
		// http2PostMultiple("http://127.0.0.2:8000/test-server/PostUser")
		http2PostMultipleThroughput("http://127.0.0.2:8000/test-server/PostUser")
		time.Sleep(2 * time.Second)
	}

	f, err := os.OpenFile("/home/johnson/onvm/result/http.csv", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 6666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	w := csv.NewWriter(f)
	defer f.Close()
	fmt.Println(datas)
	w.Write(datas)
	w.Flush()

	time.Sleep(20 * time.Second)
}

func httpGET(url string) {
	logger.Log.Warnln("Using http1")
	var err error

	t1 := time.Now()
	res, err := http.Get(url)
	rsp, err := ioutil.ReadAll(res.Body)
	t2 := time.Now()

	delay := t2.Sub(t1).Seconds() * 1000
	defer res.Body.Close()

	if err != nil {
		logger.Log.Fatal(err)
	}

	logger.Log.Debugf("URL: %s\nResponse: %s\nDelay: %f", url, rsp, delay)
	datas = append(datas, fmt.Sprintf("%.8f", delay))

	return
}

func httpPost(url string) {
	logger.Log.Warnln("Using http1")
	var err error
	user := &User{
		Id:       strconv.Itoa(ID),
		Name:     fmt.Sprintf("Ben%d", ID),
		Password: "qqqqq",
	}
	ID = ID + 1
	b, err := json.Marshal(user)

	t1 := time.Now()
	res, err := http.Post(url, "application/json", bytes.NewReader(b))
	rsp, err := ioutil.ReadAll(res.Body)
	t2 := time.Now()

	delay := t2.Sub(t1).Seconds() * 1000
	defer res.Body.Close()

	if err != nil {
		logger.Log.Fatal(err)
	}

	logger.Log.Debugf("URL: %s\nResponse: %s\nDelay: %f", url, rsp, delay)
	datas = append(datas, fmt.Sprintf("%.8f", delay))

	return
}

func http2Post(url string) {
	var err error
	user := &User{
		Id:       strconv.Itoa(ID),
		Name:     fmt.Sprintf("Ben%d", ID),
		Password: "qqqqq",
	}
	ID = ID + 1
	b, err := json.Marshal(user)

	var t http.RoundTripper

	if !USE_ONVM_TRANSPORT {
		if USE_ONVM {
			logger.Log.Warnln("Using ONVM socket")
			t = &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return onvmpoller.DialONVM("onvm", addr)
				},
			}
		} else {
			logger.Log.Warnln("Using TCP socket")
			t = &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial("tcp", addr)
				},
			}
		}
	} else {
		logger.Log.Warnln("Using http2.OnvmTransport")
		t = &http2.OnvmTransport{
			UseONVM: true,
		}
	}

	c := &http.Client{
		Transport: t,
	}

	t1 := time.Now()
	resp, err := c.Post(url, "application/json", bytes.NewReader(b))
	rsp, _ := ioutil.ReadAll(resp.Body)
	t2 := time.Now()

	delay := t2.Sub(t1).Seconds() * 1000
	if err != nil {
		logger.Log.Fatal("request error")
	}
	defer resp.Body.Close()

	logger.Log.Debugf("URL: %s\n\u001b[92m[Response]\u001b[0m \n%s\n\u001b[95m[Delay]\u001b[0m %f", url, rsp, delay)
	datas = append(datas, fmt.Sprintf("%.8f", delay))
}

func http2GET(url string) {
	var err error
	var t http.RoundTripper

	if !USE_ONVM_TRANSPORT {
		logger.Log.Warnln("Using http2.Transport")
		if USE_ONVM {
			logger.Log.Warnln("Using ONVM socket")
			t = &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return onvmpoller.DialONVM("onvm", addr)
				},
			}
		} else {
			logger.Log.Warnln("Using TCP socket")
			t = &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial("tcp", addr)
				},
			}
		}
	} else {
		logger.Log.Warnln("Using http2.OnvmTransport")
		t = &http2.OnvmTransport{
			UseONVM: true,
		}
	}

	c := &http.Client{
		Transport: t,
	}

	t1 := time.Now()
	res, err := c.Get(url)
	rsp, err := ioutil.ReadAll(res.Body)
	t2 := time.Now()

	delay := t2.Sub(t1).Seconds() * 1000
	defer res.Body.Close()

	if err != nil {
		logger.Log.Fatal(err)
	}

	logger.Log.Debugf("URL: %s\n\u001b[92m[Response]\u001b[0m \n%s\n\u001b[95m[Delay]\u001b[0m %f", url, rsp, delay)
	datas = append(datas, fmt.Sprintf("%.8f", delay))
}

func http2PostMultiple(url string) {
	// var err error
	var t http.RoundTripper

	if !USE_ONVM_TRANSPORT {
		if USE_ONVM {
			logger.Log.Warnln("Using ONVM socket")
			t = &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return onvmpoller.DialONVM("onvm", addr)
				},
			}
		} else {
			logger.Log.Warnln("Using TCP socket")
			t = &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial("tcp", addr)
				},
			}
		}
	} else {
		logger.Log.Warnln("Using http2.OnvmTransport")
		t = &http2.OnvmTransport{
			UseONVM: true,
		}
	}

	c := &http.Client{
		Transport: t,
	}

	for i := 0; i < 200; i++ {
		user := &User{
			Id:       strconv.Itoa(ID),
			Name:     fmt.Sprintf("Ben%d", ID),
			Password: "qqqqq",
		}
		ID = ID + 1
		b, err := json.Marshal(user)

		t1 := time.Now()
		resp, err := c.Post(url, "application/json", bytes.NewReader(b))
		ioutil.ReadAll(resp.Body)
		t2 := time.Now()

		delay := t2.Sub(t1).Seconds() * 1000
		if err != nil {
			logger.Log.Fatal("request error")
		}
		datas = append(datas, fmt.Sprintf("%.8f", delay))

		// logger.Log.Infof("URL: %s\n\u001b[92m[Response]\u001b[0m \n%s\n\u001b[95m[Delay]\u001b[0m %f", url, rsp, delay)
		resp.Body.Close()
	}
}

func http2PostMultipleThroughput(url string) {
	// var err error
	var t http.RoundTripper

	if !USE_ONVM_TRANSPORT {
		if USE_ONVM {
			logger.Log.Warnln("Using ONVM socket")
			t = &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return onvmpoller.DialONVM("onvm", addr)
				},
			}
		} else {
			logger.Log.Warnln("Using TCP socket")
			t = &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial("tcp", addr)
				},
			}
		}
	} else {
		logger.Log.Warnln("Using http2.OnvmTransport")
		t = &http2.OnvmTransport{
			UseONVM: true,
		}
	}

	bytes_ch := make(chan int, n_thread+1)
	wg := new(sync.WaitGroup)

	wg.Add(n_thread)
	t1 := time.Now()
	timeout := t1.Add(1 * time.Second)
	for i := 0; i < n_thread; i++ {
		go func(id int) {
			c := &http.Client{
				Transport: t,
			}
			sum := 0
			for x := 0; x < n_req; x++ {
				id = id*n_req + x
				user := &User{
					Id:       strconv.Itoa(id),
					Name:     fmt.Sprintf("Ben%d", id),
					Password: "qqqqq",
				}
				b, _ := json.Marshal(user)

				_, rerr := c.Post(url, "application/json", bytes.NewReader(b))
				if rerr == nil {
					sum++
				}

				if time.Now().After(timeout) {
					break
				}
			}
			bytes_ch <- sum
			wg.Done()
		}(i)
	}
	// wg.Wait()

	sum := 0
	counter := 0
	var lateny_sum float64 = 0.0
	for v := range bytes_ch {
		sum += v
		lateny_sum += 1000000.0 / float64(v)
		counter++
		if counter == n_thread {
			break
		}
	}

	fmt.Printf("Register user: %v, Latency: %v\n", sum, lateny_sum/float64(n_thread))
}
