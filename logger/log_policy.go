package logger

import (
	"golang.org/x/time/rate"
	"math/rand"
)

const (
	defaultR     = 10
	defaultB     = 1
	defaultRatio = 0.1
)

type Policy interface {
	Allow() bool
}

// RateLimitPolicy rate limit to control log print
// r: rate, b: burst
type RateLimitPolicy struct {
	rateLimiter *rate.Limiter
}

// NewRateLimitPolicy create a rate limit policy
// r: rate, b: burst
// example: NewRateLimitPolicy(10, 1) means 10 logs per second
func NewRateLimitPolicy(r float64, b int) Policy {
	if r <= 0 || b <= 0 {
		return &RateLimitPolicy{
			rateLimiter: rate.NewLimiter(defaultR, defaultB),
		}
	}

	return &RateLimitPolicy{
		rateLimiter: rate.NewLimiter(rate.Limit(r), b),
	}
}

func (rp *RateLimitPolicy) Allow() bool {
	return rp.rateLimiter.Allow()
}

// SamplingPolicy log print sampling with ratio
type SamplingPolicy struct {
	sampleRatio float64
}

func NewSamplingPolicy(ratio float64) Policy {
	if ratio <= 0 {
		return &SamplingPolicy{
			sampleRatio: defaultRatio,
		}
	}

	return &SamplingPolicy{
		sampleRatio: ratio,
	}
}

func (sp *SamplingPolicy) Allow() bool {
	if rand.Float64() < sp.sampleRatio {
		return true
	}
	return false
}

// AllowAllPolicy allow all log print
type AllowAllPolicy struct{}

func NewAllowAllPolicy() Policy {
	return &AllowAllPolicy{}
}

func (ap *AllowAllPolicy) Allow() bool {
	return true
}

// RejectAllPolicy reject all log print
type RejectAllPolicy struct{}

func NewRejectAllPolicy() Policy {
	return &RejectAllPolicy{}
}

func (rp *RejectAllPolicy) Allow() bool {
	return false
}
