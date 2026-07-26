// Example: httpext/v3 — interceptor chain on the standard *http.Client.
//
// Wires traffic + metrics + header injection onto an *http.Client via
// Interceptor.Apply, then uses the standard library client as normal. Point
// it at any HTTP server (a refused connection still demonstrates the traffic
// log on the error path).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/tenz-io/gokit/httpext/v3"
	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/monitor/v3"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithFilePath("log"),
		logger.WithCaller(true),
		logger.WithCallerSkip(1),
		logger.WithTraffic(true),
	)
}

func main() {
	// Inject a monitor Exporter so WithEnableMetrics records something. At the
	// request edge this is the single-flight injection; downstream calls reuse it.
	ctx := monitor.Init(context.Background(), "httpext-example")

	interceptor := httpext.NewInterceptorWithOpts(
		httpext.WithEnableMetrics(true),
		httpext.WithEnableTraffic(true),
		httpext.WithHeaders(map[string]string{
			httpext.HeaderNameAuthorization: "Bearer token",
		}),
	)

	httpCli := &http.Client{Timeout: 5 * time.Second}
	interceptor.Apply(httpCli)

	// 1. POST a JSON request using the standard library client directly.
	type Req struct {
		Name string `json:"name"`
	}
	reqBody, _ := json.Marshal(&Req{Name: "gokit"})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://example.com/api", bytes.NewReader(reqBody))
	req.Header.Set(httpext.HeaderNameContentType, "application/json")

	resp, err := httpCli.Do(req)
	log.Printf("post err: %v, status: %v", err, statusOf(resp))

	// 2. GET via the standard library shorthand; the chain runs transparently.
	getResp, gerr := httpCli.Get("https://example.com/missing")
	log.Printf("get err: %v, status: %v", gerr, statusOf(getResp))

	time.Sleep(200 * time.Millisecond) // let traffic.log flush
}

// statusOf returns the response's status code, or 0 when there is no response
// (e.g. a transport error).
func statusOf(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}
