package main

import (
	"fmt"
	// "net/http"
	"os"
	"time"

	"testbed/http/http_server/logger"
	"testbed/http/http_server/router"

	"github.com/nycu-ucr/gonet/http"
	"github.com/nycu-ucr/net/http2"
	"github.com/nycu-ucr/net/http2/h2c"
	"github.com/nycu-ucr/onvmpoller"
	"github.com/sirupsen/logrus"
)

const IP = "127.0.0.2"
const port = "8000"
const USE_ONVM = true

var host = IP + ":" + port

func checkErr(err error) {
	if err == nil {
		return
	}
	logger.ServerLog.Errorln("While listening", err)
	os.Exit(1)
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
	onvmpoller.SetLocalAddress(IP)

	fmt.Printf("Server started")
	logger.SetLogLevel(logrus.TraceLevel)
	test_router := router.NewGin()
	router.AddService(test_router)

	server := H2CServerUpgrade(test_router)
	logger.ServerLog.Infoln("Listening on", host)
	checkErr(server.ListenAndServe())
}

// This server supports "H2C upgrade" and "H2C prior knowledge" along with
// standard HTTP/2 and HTTP/1.1 that golang natively supports.
func H2CServerUpgrade(handler http.Handler) *http.Server {
	h2s := &http2.Server{
		// TODO: extends the idle time after re-use openapi client
		IdleTimeout: 10 * time.Second,
	}

	// handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, %v, http: %v", r.URL.Path, r.TLS == nil)
	// })

	server := &http.Server{
		USING_ONVM_SOCKET: false,
		Addr:              host,
		Handler:           h2c.NewHandler(handler, h2s),
		// Handler: onvm2c.NewHandler(handler, h2s),
	}

	return server
}
