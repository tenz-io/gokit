package httpcli

import "time"

var (
	defaultConfig = Config{
		EnableTraffic: true,
		EnableMetrics: false,
		Headers:       nil,
		SlowLogFloor:  0,
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
	// SlowLogFloor is the threshold for slow log,
	// if the request duration is greater than this value, it will be logged as slow log
	// if 0, it will not log slow log
	SlowLogFloor time.Duration `json:"slow_log_floor" yaml:"slow_log_floor"`
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

func WithSlowLogFloor(slowLogFloor time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.SlowLogFloor = slowLogFloor
	}
}
