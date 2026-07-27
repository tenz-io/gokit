package httpext

// defaultConfig 在未提供任何 option 时应用。Traffic 默认开启,
// 因此开箱即用的 client 会记录 request/response span;metrics 默认关闭,
// 因为 Prometheus 必须在进程级别配置 (monitor/v3)。
var defaultConfig = Config{
	EnableTraffic: true,
	EnableMetrics: false,
	Headers:       nil,
}

// Config 控制 interceptor chain 接入哪些 transport。
type Config struct {
	// EnableTraffic 通过 logger/v3 的 traffic logger 开启 request/response
	// traffic 日志。当 request context 携带 tracer.FlagDebug 时,无论此 flag
	// 取值如何,也会自动记录 traffic。
	EnableTraffic bool `json:"enable_traffic" yaml:"enable_traffic"`
	// EnableMetrics 通过 monitor/v3 开启每请求的 latency/counter 记录。
	// 默认关闭:它要求在 context 中注入 Exporter (见 monitor.Init)。
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`
	// Headers 通过 Header.Set 注入到每个出站 request,因此同名 per-request
	// header 会覆盖已配置的值。
	// 示例:{"Authorization": "Bearer token"}。
	Headers map[string]string `json:"headers" yaml:"headers"`
}

// ConfigOption 变更 Config。NewInterceptorWithOpts 将其叠加在
// defaultConfig 之上。
type ConfigOption func(cfg *Config)

// WithEnableTraffic 启用或禁用 traffic 日志。
func WithEnableTraffic(enableTraffic bool) ConfigOption {
	return func(cfg *Config) {
		cfg.EnableTraffic = enableTraffic
	}
}

// WithEnableMetrics 启用或禁用 Prometheus metrics 记录。
func WithEnableMetrics(enableMetrics bool) ConfigOption {
	return func(cfg *Config) {
		cfg.EnableMetrics = enableMetrics
	}
}

// WithHeaders 设置注入到每个出站 request 的 header。
func WithHeaders(headers map[string]string) ConfigOption {
	return func(cfg *Config) {
		cfg.Headers = headers
	}
}
