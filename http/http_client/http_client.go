package main

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/gonet/http"
	"github.com/nycu-ucr/net/http2"
	"github.com/nycu-ucr/onvmpoller"
	"github.com/sirupsen/logrus"
)

const (
	EPOCHS             = 1
	USE_ONVM           = true
	USE_ONVM_TRANSPORT = true
)

var (
	ID    int
	datas []string
)

type User struct {
	Id       string `json:"UserId"`
	Name     string `json:"UserName"`
	Password string `json:"UserPassword"`
}

func main() {
	if USE_ONVM {
		go func() {
			time.Sleep(30 * time.Second)
			onvmpoller.CloseONVM()
			os.Exit(1)
		}()
		defer onvmpoller.CloseONVM()
	}
	onvmpoller.SetLocalAddress("127.0.0.1")

	/* Init global var */
	logger.Log.Logger.SetLevel(logrus.InfoLevel)
	ID = 50
	datas = make([]string, 0)

	for i := 0; i < EPOCHS; i++ {
		http2Post("http://127.0.0.2:8000/test-server/PostUser")
		http2GET("http://127.0.0.2:8000/test-server/GetUser")
		http2Post("http://127.0.0.2:8000/test-server/PostUser")
		http2GET("http://127.0.0.2:8000/test-server/GetUser")
		http2GET("http://127.0.0.2:8000/test-server/GetUser/1")
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

	time.Sleep(10 * time.Second)
}

func httpGET(url string) {
	logger.Log.Warnln("Using http1")
	var err error

	t1 := time.Now()
	res, err := http_clients[hc_idx].Get(url)
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
	res, err := http_clients[hc_idx].Post(url, "application/json", bytes.NewReader(b))
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
		logger.Log.Warnln("Using http2.Transport")
		t = &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return onvmpoller.DialONVM("onvm", addr)
			},
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
		t = &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return onvmpoller.DialONVM("onvm", addr)
			},
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
