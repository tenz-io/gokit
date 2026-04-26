package retriever

import "time"

var defaultConfig = Config{
	MaxAttempt: 3,
	Backoff:    NewExponentialBackoff(100, 2.0, 0.3),
}

// Config holds retriever parameters. Zero values use sensible defaults.
type Config struct {
	MaxAttempt          int
	MaxTotalAttemptTime time.Duration
	Backoff             Backoff
}

// ConfigOpt is a functional option for configuring a Retriever.
type ConfigOpt func(*Config)

// WithMaxAttempt sets the maximum number of attempts (including the first call).
func WithMaxAttempt(n int) ConfigOpt {
	return func(c *Config) { c.MaxAttempt = n }
}

// WithMaxTotalAttemptTime sets a deadline across all attempts.
func WithMaxTotalAttemptTime(d time.Duration) ConfigOpt {
	return func(c *Config) { c.MaxTotalAttemptTime = d }
}

// WithBackoff sets the backoff strategy.
func WithBackoff(b Backoff) ConfigOpt {
	return func(c *Config) { c.Backoff = b }
}
