package dbtracker

import "time"

var (
	defaultConfig = Config{
		EnableTraffic:  true,
		EnableMetrics:  false,
		EnableErrorLog: false,
		SlowLogFloor:   5 * time.Second,
	}
)

type Config struct {
	// EnableTraffic is a flag to enable traffic interceptor.
	EnableTraffic bool `yaml:"enable_traffic" json:"enable_traffic"`
	// EnableMetrics is a flag to enable metrics interceptor.
	EnableMetrics bool `yaml:"enable_metrics" json:"enable_metrics"`
	// EnableErrorLog is a flag to enable error log interceptor.
	EnableErrorLog bool `yaml:"enable_error_log" json:"enable_error_log"`
	// SlowLogFloor is the minimum duration to log slow query.
	// when the query duration is larger than this value, it will be logged.
	// if 0, it will not log slow query.
	SlowLogFloor time.Duration `yaml:"slow_log_floor" json:"slow_log_floor"`
}

type ConfigOption func(*Config)

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

func WithSlowLogFloor(floor time.Duration) ConfigOption {
	return func(c *Config) {
		c.SlowLogFloor = floor
	}
}

func WithErrorLog(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableErrorLog = enable
	}
}
