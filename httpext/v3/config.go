package httpext

// defaultConfig is applied when no options are given. Traffic is on by default
// so an out-of-the-box client records request/response spans; metrics is off
// because Prometheus must be configured at the process level (monitor/v3).
var defaultConfig = Config{
	EnableTraffic: true,
	EnableMetrics: false,
	Headers:       nil,
}

// Config controls which transports the interceptor chain wires up.
type Config struct {
	// EnableTraffic turns on request/response traffic logging via logger/v3's
	// traffic logger. Traffic is also recorded automatically when the request
	// context carries tracer.FlagDebug, regardless of this flag.
	EnableTraffic bool `json:"enable_traffic" yaml:"enable_traffic"`
	// EnableMetrics turns on per-request latency/counter recording via
	// monitor/v3. Off by default: it requires an Exporter injected into the
	// context (see monitor.Init).
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`
	// Headers are injected into every outbound request via Header.Set, so a
	// per-request header of the same name overrides the configured one.
	// Example: {"Authorization": "Bearer token"}.
	Headers map[string]string `json:"headers" yaml:"headers"`
}

// ConfigOption mutates a Config. NewInterceptorWithOpts layers these over
// defaultConfig.
type ConfigOption func(cfg *Config)

// WithEnableTraffic enables or disables traffic logging.
func WithEnableTraffic(enableTraffic bool) ConfigOption {
	return func(cfg *Config) {
		cfg.EnableTraffic = enableTraffic
	}
}

// WithEnableMetrics enables or disables Prometheus metrics recording.
func WithEnableMetrics(enableMetrics bool) ConfigOption {
	return func(cfg *Config) {
		cfg.EnableMetrics = enableMetrics
	}
}

// WithHeaders sets the headers injected into every outbound request.
func WithHeaders(headers map[string]string) ConfigOption {
	return func(cfg *Config) {
		cfg.Headers = headers
	}
}
