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

// startTime marks process startup for /ping uptime reporting. It is read-only
// after package init.
var startTime = time.Now()

// AddPrometheusHandler mounts the Prometheus metrics handler at /metrics.
func AddPrometheusHandler(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
}

// AddPingHandler mounts the liveness handler at /ping.
func AddPingHandler(mux *http.ServeMux) {
	mux.HandleFunc("/ping", PingHandler)
}

// AddProfilingHandler mounts net/http/pprof endpoints under /debug/pprof/.
// Unlike v2 it always registers (v2 skipped registration on the global
// DefaultServeMux, which silently dropped profiling in the common path).
func AddProfilingHandler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// PingHandler reports hostname, start time and uptime. Used by load balancers
// and orchestrators for liveness checks.
func PingHandler(w http.ResponseWriter, _ *http.Request) {
	hostname, _ := os.Hostname()
	_, _ = io.WriteString(w, fmt.Sprintf(
		"%s | StartAt: %s | Uptime: %s\n",
		hostname,
		startTime.Truncate(time.Second),
		time.Since(startTime).Truncate(time.Second),
	))
}
