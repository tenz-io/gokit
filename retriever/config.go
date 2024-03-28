package retriever

import "time"

var (
	defaultConfig = Config{
		MaxAttempt:          3,
		MaxTotalAttemptTime: 0,
		Backoff:             NewExponentialBackoff(100, 2.0, 0.3),
		UseGoroutine:        false,
	}
)

// ConfigOpt is a function to set the Config
type ConfigOpt func(*Config)

// Config for retriever, all fields are optional
type Config struct {
	// MaxAttempt Maximum number of attempt before failing (default: 3)
	MaxAttempt int
	// MaxTotalAttemptTime Maximum total duration to wait for result (default: unlimited/until context expired)
	MaxTotalAttemptTime time.Duration
	// Backoff An implementation of Backoff to calculate the time duration to wait for the next operation.
	// Will default to ExponentialBackoff with 100ms base and 2.0 factor
	Backoff Backoff
	// UseGoroutine is a flag to determine whether DoFunc will be executed using goroutine or not.
	// If true, DoFunc will be executed using goroutine.
	// This allows for long function without context support (default: false)
	UseGoroutine bool
}

// WithMaxAttempt sets the MaxAttempt field in Config
func WithMaxAttempt(maxAttempt int) ConfigOpt {
	return func(c *Config) {
		c.MaxAttempt = maxAttempt
	}
}

// WithMaxTotalAttemptTime sets the MaxTotalAttemptTime field in Config
func WithMaxTotalAttemptTime(maxTotalAttemptTime time.Duration) ConfigOpt {
	return func(c *Config) {
		c.MaxTotalAttemptTime = maxTotalAttemptTime
	}
}

// WithBackoff sets the Backoff field in Config
func WithBackoff(backoff Backoff) ConfigOpt {
	return func(c *Config) {
		c.Backoff = backoff
	}
}

// WithUseGoroutine sets the UseGoroutine field in Config
func WithUseGoroutine(useGoroutine bool) ConfigOpt {
	return func(c *Config) {
		c.UseGoroutine = useGoroutine
	}
}
