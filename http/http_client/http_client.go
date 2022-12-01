package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	// "net/http"
	"os"
	"strconv"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/gonet/http"
	"github.com/nycu-ucr/net/http2"
	"github.com/nycu-ucr/onvmpoller"
)

const (
	EPOCHS             = 1
	USE_ONVM           = true
	USE_ONVM_TRANSPORT = true
)

var (
	ID int
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
	onvmpoller.SetLocalAddress("127.0.0.3")

	/* Init global var */
	ID = 50

	for i := 0; i < EPOCHS; i++ {
		http2Post("http://127.0.0.2:8000/test-server/PostUser")
		http2GET("http://127.0.0.2:8000/test-server/GetUser")
		http2Post("http://127.0.0.2:8000/test-server/PostUser")
		http2GET("http://127.0.0.2:8000/test-server/GetUser")
		http2GET("http://127.0.0.2:8000/test-server/GetUser/1")
	}

	time.Sleep(10 * time.Second)
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

	logger.Log.Infof("URL: %s\nResponse: %s\nDelay: %f", url, rsp, delay)

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

	logger.Log.Infof("URL: %s\nResponse: %s\nDelay: %f", url, rsp, delay)

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

	logger.Log.Infof("URL: %s\n\u001b[92m[Response]\u001b[0m \n%s\n\u001b[95m[Delay]\u001b[0m %f", url, rsp, delay)
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

	logger.Log.Infof("URL: %s\n\u001b[92m[Response]\u001b[0m \n%s\n\u001b[95m[Delay]\u001b[0m %f", url, rsp, delay)

	return
}
