package grpcext

var (
	defaultConfig = Config{
		EnabledServerTraffic: false,
		EnabledClientTraffic: false,
		EnabledServerMetrics: false,
		EnabledClientMetrics: false,
	}
)

type Config struct {
	EnabledServerTraffic bool
	EnabledClientTraffic bool
	EnabledServerMetrics bool
	EnabledClientMetrics bool
}

type Option func(*Config)

func WithServerTraffic(enabled bool) Option {
	return func(c *Config) {
		c.EnabledServerTraffic = enabled
	}
}

func WithClientTraffic(enabled bool) Option {
	return func(c *Config) {
		c.EnabledClientTraffic = enabled
	}
}

func WithServerMetrics(enabled bool) Option {
	return func(c *Config) {
		c.EnabledServerMetrics = enabled
	}
}

func WithClientMetrics(enabled bool) Option {
	return func(c *Config) {
		c.EnabledClientMetrics = enabled
	}
}
