package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nycu-ucr/onvmpoller"
)

func main() {
	go func() {
		time.Sleep(30 * time.Second)
		onvmpoller.CloseONVM()
		os.Exit(1)
	}()
	defer onvmpoller.CloseONVM()

	res, err := http.Get("http://127.0.0.2:8000/test-server/GetUser")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	sitemap, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", sitemap)

	res, err = http.Get("http://127.0.0.2:8000/test-server/GetUser")
	sitemap, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", sitemap)

	res, err = http.Get("http://127.0.0.2:8000/test-server/")
	sitemap, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", sitemap)

	time.Sleep(10 * time.Second)
}
