package cache

var (
	defaultConfig = Config{
		EnableTraffic: false,
		EnableMetrics: true,
	}
)

type Config struct {
	EnableTraffic bool
	EnableMetrics bool
}

type ConfigOption func(*Config)

func WithEnableTraffic(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableTraffic = enable
	}
}

func WithEnableMetrics(enable bool) ConfigOption {
	return func(c *Config) {
		c.EnableMetrics = enable
	}
}
