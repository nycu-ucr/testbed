package main

import (
	"fmt"
	// "net"
	"net/http"
	"os"
	"time"

	"testbed/server/logger"
	"testbed/server/router"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const IP = "10.10.0.72"
const port = "5478"

var host = IP + ":" + port

func checkErr(err error) {
	if err == nil {
		return
	}
	logger.ServerLog.Errorln("While listening", err)
	os.Exit(1)
}

func main() {
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
		IdleTimeout: 1 * time.Millisecond,
	}

	// handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, %v, http: %v", r.URL.Path, r.TLS == nil)
	// })

	server := &http.Server{
		Addr:    host,
		Handler: h2c.NewHandler(handler, h2s),
	}

	return server
}
