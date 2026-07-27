package httpext

import "net/http"

// Interceptor 在 client 已有的 transport 之上组合一组 http.RoundTripper
// 层 (header injection、metrics、traffic)。使用 NewInterceptor 或
// NewInterceptorWithOpts 构建后,通过 Apply 应用到 *http.Client。
type Interceptor interface {
	// Intercept 返回一个 http.RoundTripper,按 newTransporters 顺序用每个
	// 激活层包装 tripper。nil tripper 回退到 http.DefaultTransport。
	Intercept(tripper http.RoundTripper) http.RoundTripper
	// Apply 用已拦截的 chain 替换 hc.Transport。nil hc 为 no-op。
	Apply(hc *http.Client)
}

type interceptor struct {
	config Config
}

// NewInterceptorWithOpts 由函数式 option 构建 interceptor,叠加在
// defaultConfig 之上。
func NewInterceptorWithOpts(opts ...ConfigOption) Interceptor {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	return NewInterceptor(config)
}

// NewInterceptor 由显式 Config 构建 interceptor。
func NewInterceptor(config Config) Interceptor {
	return &interceptor{
		config: config,
	}
}

func (i *interceptor) Intercept(tripper http.RoundTripper) http.RoundTripper {
	transport := tripper
	if transport == nil {
		transport = http.DefaultTransport
	}

	// 由内向外 fold chain:每个激活的 transporter 包装运行中的
	// transport,因此结果为 traffic(metrics(injectHeader(parent)))。
	// 未激活的层被跳过,因此 chain 仅包含已启用的层。
	for _, newTransporter := range newTransporters {
		layer := newTransporter(i.config, transport)
		if layer.active() {
			transport = layer
		}
	}

	return transport
}

func (i *interceptor) Apply(hc *http.Client) {
	if hc == nil {
		return
	}

	hc.Transport = i.Intercept(hc.Transport)
}
