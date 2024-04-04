package httpcli

var (
	defaultConfig = Config{
		EnableTraffic: true,
		EnableMetrics: false,
		Headers:       nil,
	}
)

type Config struct {
	// EnableTraffic enables traffic logging
	EnableTraffic bool `json:"enable_traffic" yaml:"enable_traffic"`
	// EnableMetrics enables metrics prometheus
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`
	// Headers is a map of headers to inject into the request
	// The key is the header name and the value is the header value
	// Example: {"Authorization": "Bearer token", "Content-Type": "application/json"}
	Headers map[string]string `json:"headers" yaml:"headers"`
}

type ConfigOption func(cfg *Config)

func WithEnableTraffic(enableTraffic bool) ConfigOption {
	return func(cfg *Config) {
		cfg.EnableTraffic = enableTraffic
	}
}

func WithEnableMetrics(enableMetrics bool) ConfigOption {
	return func(cfg *Config) {
		cfg.EnableMetrics = enableMetrics
	}
}

func WithHeaders(headers map[string]string) ConfigOption {
	return func(cfg *Config) {
		cfg.Headers = headers
	}
}
