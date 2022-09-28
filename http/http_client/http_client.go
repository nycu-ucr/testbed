package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testbed/logger"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

const (
	EPOCHS = 1
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
	go func() {
		time.Sleep(30 * time.Second)
		onvmpoller.CloseONVM()
		os.Exit(1)
	}()
	defer onvmpoller.CloseONVM()

	/* Wait for ONVM to init */
	time.Sleep(5 * time.Second)

	/* Init global var */
	ID = 50

	for i := 0; i < EPOCHS; i++ {
		httpPost("http://127.0.0.2:8000/test-server/PostUser")
		httpGET("http://127.0.0.2:8000/test-server/GetUser")
		httpPost("http://127.0.0.2:8000/test-server/PostUser")
		httpGET("http://127.0.0.2:8000/test-server/GetUser")
		httpGET("http://127.0.0.2:8000/test-server/GetUser/1")
	}

	time.Sleep(10 * time.Second)
}

func httpGET(url string) {
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
