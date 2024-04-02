package httpcli

var (
	defaultConfig = Config{
		EnableTraffic: true,
		EnableMetrics: false,
	}
)

type Config struct {
	// EnableTraffic enables traffic logging
	EnableTraffic bool `json:"enable_traffic" yaml:"enable_traffic"`
	// EnableMetrics enables metrics prometheus
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`
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
