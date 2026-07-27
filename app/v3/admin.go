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

// startTime 标记进程启动时刻,用于 /ping 的 uptime 报告。在
// package init 之后为只读。
var startTime = time.Now()

// AddPrometheusHandler 将 Prometheus metrics handler 挂载到 /metrics。
func AddPrometheusHandler(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
}

// AddPingHandler 将 liveness handler 挂载到 /ping。
func AddPingHandler(mux *http.ServeMux) {
	mux.HandleFunc("/ping", PingHandler)
}

// AddProfilingHandler 将 net/http/pprof 端点挂载到 /debug/pprof/ 下。
// 与 v2 不同,它总是注册(v2 在全局
// DefaultServeMux 上跳过注册,在常见路径下静默丢失 profiling)。
func AddProfilingHandler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// PingHandler 报告 hostname、启动时间与 uptime。供 load balancer
// 与 orchestrator 用于 liveness 检查。
func PingHandler(w http.ResponseWriter, _ *http.Request) {
	hostname, _ := os.Hostname()
	_, _ = io.WriteString(w, fmt.Sprintf(
		"%s | StartAt: %s | Uptime: %s\n",
		hostname,
		startTime.Truncate(time.Second),
		time.Since(startTime).Truncate(time.Second),
	))
}
