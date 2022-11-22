package main

import (
	"bytes"
	"crypto/tls"
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
)

const (
	EPOCHS        = 1
	USE_ONVM      = true
	USE_HTTP1     = 0
	USE_HTTP2     = 1
	USE_HTTP_ONVM = 2
)

var (
	ID           int
	tcp_h1c      *http.Client
	tcp_h2c      *http.Client
	onvm_h2c     *http.Client
	http_clients [3]*http.Client
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

	tcp_h1c = http.DefaultClient
	tcp_h2c = &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS:   func(network, addr string, cfg *tls.Config) (net.Conn, error) { return net.Dial(network, addr) },
		},
	}
	onvm_h2c = &http.Client{
		Transport: &http2.OnvmTransport{
			UseONVM: true,
			// USE_ONVM: false,
		},
	}
	http_clients[USE_HTTP1] = tcp_h1c
	http_clients[USE_HTTP2] = tcp_h2c
	http_clients[USE_HTTP_ONVM] = onvm_h2c

	/* Init global var */
	ID = 50
	hc_idx := USE_HTTP2

	for i := 0; i < EPOCHS; i++ {
		httpPost("http://127.0.0.2:8000/test-server/PostUser", hc_idx)
		httpGET("http://127.0.0.2:8000/test-server/GetUser", hc_idx)
		httpPost("http://127.0.0.2:8000/test-server/PostUser", hc_idx)
		httpGET("http://127.0.0.2:8000/test-server/GetUser", hc_idx)
		httpGET("http://127.0.0.2:8000/test-server/GetUser/1", hc_idx)
	}

	time.Sleep(10 * time.Second)
}

func httpGET(url string, hc_idx int) {
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

	logger.Log.Infof("URL: %s\nResponse: %s\nDelay: %f", url, rsp, delay)

	return
}

func httpPost(url string, hc_idx int) {
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

	logger.Log.Infof("URL: %s\nResponse: %s\nDelay: %f", url, rsp, delay)

	return
}
