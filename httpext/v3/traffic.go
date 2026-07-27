package httpext

import (
	"fmt"
	"net/http"

	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/tracer/v3"
)

// trafficTransport 通过 logger/v3 的 traffic logger 记录 request/response
// span。它在 EnableTraffic 被设置,或 request context 携带
// tracer.FlagDebug (per-request debug traffic) 时记录,因此运维方可以在
// 不修改 client config 的情况下为单个 request 切换捕获。
type trafficTransport struct {
	enable  bool
	tripper http.RoundTripper
}

func newTrafficTransport(config Config, parent http.RoundTripper) transporter {
	return &trafficTransport{
		enable:  config.EnableTraffic,
		tripper: parent,
	}
}

func (tt *trafficTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	ctx := req.Context()

	// 门控:仅在配置开启,或 request context 处于 debug 模式
	// (per-request opt-in) 时捕获。关闭时,本层对 parent 为纯 pass through。
	if !tt.enable && !tracer.FromContext(ctx).IsDebug() {
		return tt.tripper.RoundTrip(req)
	}

	// 以 "send" 类型启动 traffic span (这是一次出站调用)。span 记录
	// cmd/cost/code 以及下方结构化字段;它不会读取或记录 request/response
	// body —— 保持 traffic 日志开销低且对 body 流无副作用。
	rec := logger.FromContext(ctx).StartTraffic(req.URL.Path).WithTyp(logger.TrafficTypSend)

	defer func() {
		// error/无 response 路径下的默认值:code 1 是
		// "no status known" 的哨兵值;resp_header 保持 nil。
		code := "1"
		var respHeaders http.Header
		switch {
		case err != nil:
			// 在任何 response 之前的网络/transport 错误:以 error 及
			// request 侧字段结束 span。
			rec.EndWithError(err,
				"method", req.Method,
				"url", req.URL.String(),
				"query", req.URL.Query(),
				"req_header", req.Header,
			)
			return
		case resp != nil:
			code = fmt.Sprintf("%d", resp.StatusCode)
			respHeaders = resp.Header
		}

		rec.End(nil, code,
			"method", req.Method,
			"url", req.URL.String(),
			"query", req.URL.Query(),
			"req_header", req.Header,
			"resp_header", respHeaders,
		)
	}()

	return tt.tripper.RoundTrip(req)
}

// active 报告本层是否接入 chain。trafficTransport 一旦拥有 parent 即
// 始终激活:即使 EnableTraffic 为 false,它也必须留在 chain 中,因为携带
// tracer.FlagDebug 的 per-request context 仍需能够捕获 traffic。enable flag
// 仅对非 debug request 的捕获起门控作用 (见 RoundTrip 守卫)。
func (tt *trafficTransport) active() bool {
	if tt == nil || tt.tripper == nil {
		return false
	}
	return true
}
