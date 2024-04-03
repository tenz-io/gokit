package main

import (
	syslog "log"
	"net/http"
	"time"

	"github.com/tenz-io/gokit/httpcli"
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
		logger.WithTrafficConsoleEnabled(true),
		logger.WithTrafficFileEnabled(true),
	)

}

func main() {
	interceptor := httpcli.NewInterceptorWithOpts(
		httpcli.WithEnableMetrics(true),
		httpcli.WithEnableTraffic(true),
	)

	client := http.Client{}
	interceptor.Apply(&client)

	resp, err := client.Get("http://localhost:8080")
	syslog.Printf("resp: %+v, err: %v", resp, err)

	resp, err = client.Get("https://www.google.com")
	syslog.Printf("resp: %+v, err: %v", resp, err)

	time.Sleep(100 * time.Millisecond)
}
