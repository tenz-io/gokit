package retriever

import (
	"math"
	"math/rand"
	"time"
)

// Backoff computes wait durations between retry attempts.
type Backoff interface {
	Next(failCount int) time.Duration
}

// NoBackoff returns zero wait — retries happen immediately.
type NoBackoff struct{}

func (NoBackoff) Next(_ int) time.Duration { return 0 }

// LinearBackoff returns a constant duration regardless of fail count.
type LinearBackoff struct {
	duration time.Duration
}

// NewLinearBackoff creates a LinearBackoff with the given duration in milliseconds.
func NewLinearBackoff(ms int64) Backoff {
	return &LinearBackoff{duration: time.Duration(ms) * time.Millisecond}
}

func (l *LinearBackoff) Next(_ int) time.Duration { return l.duration }

// ExponentialBackoff grows exponentially: base * (factor^failCount + jitter).
// jitter is a random component in [0, jitter) multiplied by base.
type ExponentialBackoff struct {
	base   float64 // base wait in milliseconds
	factor float64
	jitter float64
}

// NewExponentialBackoff creates an ExponentialBackoff.
// base is in milliseconds. jitter should be in [0, 1.0) (e.g. 0.3 = up to 30% jitter).
func NewExponentialBackoff(base, factor, jitter float64) Backoff {
	return &ExponentialBackoff{base: base, factor: factor, jitter: jitter}
}

func (e *ExponentialBackoff) Next(failCount int) time.Duration {
	r := 0.0
	if e.jitter > 0 {
		r = e.jitter * rand.Float64()
	}
	ms := e.base * (math.Pow(e.factor, float64(failCount)) + r)
	return time.Duration(int64(ms)) * time.Millisecond
}
