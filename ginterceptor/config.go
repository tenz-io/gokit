package ginterceptor

import "time"

var (
	defaultConfig = Config{
		EnableTracking: true,
		EnableTraffic:  true,
		EnableMetrics:  false,
		Timeout:        0,
	}
)

type Config struct {
	// EnableTraffic is a flag to enable traffic interceptor.
	EnableTraffic bool `yaml:"enable_traffic" json:"enable_traffic"`
	// EnableMetrics is a flag to enable metrics interceptor.
	EnableMetrics bool `yaml:"enable_metrics" json:"enable_metrics"`
	// EnableTracking is a flag to enable tracking interceptor.
	EnableTracking bool `yaml:"enable_tracking" json:"enable_tracking"`
	// Timeout is the maximum duration before timing out the request.
	// if 0, it will not set timeout.
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

type ConfigOption func(*Config)

func NewConfig(opts ...ConfigOption) *Config {
	cfg := defaultConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	return &cfg
}

func WithTraffic(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableTraffic = enable
	}
}

func WithMetrics(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableMetrics = enable
	}
}

func WithTimeout(timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

func WithTracking(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableTracking = enable
	}
}
