package app

import (
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	start = time.Now()
)

// AddPrometheusHandler will add handler for `/metrics` request
func AddPrometheusHandler(m *http.ServeMux) {
	m.Handle("/metrics", promhttp.Handler())
}

// AddPingHandler will add handler for `/ping` request
func AddPingHandler(m *http.ServeMux) {
	m.HandleFunc("/ping", PingHandler)
}

// AddProfilingHandler will add various pprof handler
func AddProfilingHandler(m *http.ServeMux) {
	if m == http.DefaultServeMux {
		return
	}
	m.HandleFunc("/debug/pprof/", pprof.Index)
	m.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	m.HandleFunc("/debug/pprof/profile", pprof.Profile)
	m.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	m.HandleFunc("/debug/pprof/trace", pprof.Trace)

}

// PingHandler is the http.HandlerFunc for ping request
func PingHandler(writer http.ResponseWriter, _ *http.Request) {
	hostname, _ := os.Hostname()
	_, _ = io.WriteString(
		writer,
		fmt.Sprintf(
			"%s | StartAt: %s | Uptime: %s\n",
			hostname,
			start.Truncate(time.Second),
			time.Since(start).Truncate(time.Second),
		),
	)
}
