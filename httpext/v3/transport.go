package httpext

import (
	"fmt"
	"net/http"

	"github.com/tenz-io/gokit/monitor/v3"
)

// newTransporters 是有序的 chain 构建器。顺序为最外层优先:
// injectHeader 先于 metrics,metrics 先于 traffic,因此最终的
// http.RoundTripper 为 traffic(metrics(injectHeader(parent)))。每一层仅当其
// active() 返回 true 时才接入 chain,所以禁用的层是 no-op pass through
// 而非一个死的 wrapper。
var newTransporters = []newTransporterFunc{
	newInjectHeaderTransport,
	newMetricsTransport,
	newTrafficTransport,
}

// HeaderName 常量命名本包设置或读取的 header。
const (
	HeaderNameAuthorization = "Authorization"
	HeaderNameContentType   = "Content-Type"
)

// transporter 是 chain 层满足的内部契约:它包装一个 parent RoundTripper 并
// 报告自身是否激活。未激活的层会从 chain 中剔除,因此调用方(以及测试)
// 看到的是仅由启用层构成的扁平 chain。
type transporter interface {
	http.RoundTripper
	active() bool
}

// newTransporterFunc 基于 config 在 parent 之上构建一个 chain 层。
type newTransporterFunc func(config Config, parent http.RoundTripper) transporter

// metricsTransport 通过 monitor/v3 记录每请求的 latency 和 status code 计数。
// 当 EnableMetrics 为 false,或 context 未携带 Exporter 时
// (此时 monitor.Begin 返回 nil-safe 的 Recorder),它是 no-op。
type metricsTransport struct {
	enable  bool
	tripper http.RoundTripper
}

func newMetricsTransport(config Config, parent http.RoundTripper) transporter {
	return &metricsTransport{
		enable:  config.EnableMetrics,
		tripper: parent,
	}
}

func (mt *metricsTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	var (
		ctx  = req.Context()
		url  = req.URL.Path
		code = "1" // 哨兵值:1 表示调用在状态已知之前已出错
	)

	rec := monitor.Begin(ctx, url)

	defer func() {
		if err == nil && resp != nil {
			code = fmt.Sprintf("%d", resp.StatusCode)
		}
		rec.EndWithCode(code)
	}()

	return mt.tripper.RoundTrip(req)
}

func (mt *metricsTransport) active() bool {
	if mt == nil || mt.tripper == nil {
		return false
	}
	return mt.enable
}

// injectHeaderTransport 在委托给 parent 之前为每个出站 request 设置已配置的
// header。Header.Set 是替换语义,因此即便调用方在具体 request 上设置了同名
// header 仍会被配置值覆盖 —— 仅配置在所有调用中应保持一致的 header。
type injectHeaderTransport struct {
	tripper http.RoundTripper
	headers map[string]string
}

func newInjectHeaderTransport(config Config, parent http.RoundTripper) transporter {
	return &injectHeaderTransport{
		tripper: parent,
		headers: config.Headers,
	}
}

func (it *injectHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range it.headers {
		req.Header.Set(key, value)
	}

	return it.tripper.RoundTrip(req)
}

func (it *injectHeaderTransport) active() bool {
	return it != nil && it.tripper != nil && len(it.headers) > 0
}
