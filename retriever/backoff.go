package retriever

import (
	"math/rand"
	"time"
)

// Backoff allows implementation for custom backoff strategy
type Backoff interface {
	// Next returns the duration in milliseconds to wait for the next operation
	Next(failCount int) time.Duration
}

// NoBackoff is a backoff strategy with 0 wait time.
// So the operation will be retried immediately.
type NoBackoff struct {
}

func (n *NoBackoff) Next(_ int) time.Duration {
	return 0
}

// LinearBackoff is a backoff strategy which always return the same duration.
type LinearBackoff struct {
	duration time.Duration // linear time duration in milliseconds to wait for
}

// NewLinearBackoff creates and returns new LinearBackoff
func NewLinearBackoff(duration int64) Backoff {
	return &LinearBackoff{duration: time.Duration(duration) * time.Millisecond}
}

func (l *LinearBackoff) Next(_ int) time.Duration {
	return l.duration
}

// ExponentialBackoff is a backoff strategy which increase exponentially
// over failCount
type ExponentialBackoff struct {
	base   float64 // the base wait time
	factor float64 // exponential factor
	// jitter to the backoff, random from 0.0 to jitter,
	// 0 means no jitter, 1.0 means 0% to 100% of the base
	// suggested value range from 0.1 to 0.5
	jitter float64
}

// NewExponentialBackoff creates and returns new ExponentialBackoff
// base is in milliseconds
func NewExponentialBackoff(base, factor, jitter float64) Backoff {
	return &ExponentialBackoff{
		base:   base,
		factor: factor,
		jitter: jitter,
	}
}

// Next returns the duration in milliseconds to wait for the next operation
// the formula is base * (factor^failCount + random(0, jitter))
func (e ExponentialBackoff) Next(failCount int) time.Duration {
	var r float64
	if e.jitter > 0 {
		r = e.jitter * rand.Float64() // random from 0.0 to jitter
	}
	return time.Duration(int64(e.base*(pow(e.factor, failCount)+r))) * time.Millisecond
}

// pow is a helper function to calculate the power of a number
func pow(base float64, exp int) float64 {
	if exp == 0 {
		return 1
	}

	result := base
	for i := 2; i <= exp; i++ {
		result *= base
	}
	return result
}
