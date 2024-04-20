package main

import (
	"log"
	"net/http"
	"time"

	"github.com/tenz-io/gokit/httpext"
	"github.com/tenz-io/gokit/logger"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(true),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)

	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)

}

func main() {
	interceptor := httpext.NewInterceptorWithOpts(
		httpext.WithEnableMetrics(true),
		httpext.WithEnableTraffic(true),
	)

	client := http.Client{}
	interceptor.Apply(&client)

	resp, err := client.Get("http://localhost:8080")
	log.Printf("resp: %+v, err: %v", resp, err)

	resp, err = client.Get("https://www.google.com")
	log.Printf("resp: %+v, err: %v", resp, err)

	time.Sleep(100 * time.Millisecond)
}
