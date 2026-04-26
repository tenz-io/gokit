// Package retriever provides a configurable retry mechanism with pluggable backoff strategies.
package retriever

import (
	"context"
	"fmt"
	"time"
)

// DoFunc is a retryable function. Return (result, false, err) to stop retrying on error.
type DoFunc func(ctx context.Context) (resp any, retry bool, err error)

// DoFuncAlwaysRetry is a function that is always retried on error until success or limit.
type DoFuncAlwaysRetry func(ctx context.Context) (resp any, err error)

// Retriever executes functions with retry logic.
type Retriever interface {
	Do(ctx context.Context, fn DoFunc) (any, error)
	DoAlwaysRetry(ctx context.Context, fn DoFuncAlwaysRetry) (any, error)
}

type retrier struct {
	maxAttempt          int
	maxTotalAttemptTime time.Duration
	backoff             Backoff
}

// NewRetriever creates a Retriever from a Config.
func NewRetriever(config Config) Retriever {
	if config.MaxAttempt <= 0 {
		config.MaxAttempt = 3
	}
	if config.Backoff == nil {
		config.Backoff = NewExponentialBackoff(100, 2.0, 0.3)
	}
	return &retrier{
		maxAttempt:          config.MaxAttempt,
		maxTotalAttemptTime: config.MaxTotalAttemptTime,
		backoff:             config.Backoff,
	}
}

// New creates a Retriever from functional options.
func New(opts ...ConfigOpt) Retriever {
	cfg := defaultConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewRetriever(cfg)
}

func (r *retrier) DoAlwaysRetry(ctx context.Context, fn DoFuncAlwaysRetry) (any, error) {
	return r.Do(ctx, func(ctx context.Context) (any, bool, error) {
		result, err := fn(ctx)
		return result, true, err
	})
}

func (r *retrier) Do(ctx context.Context, fn DoFunc) (any, error) {
	if fn == nil {
		return nil, fmt.Errorf("retriever: fn is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("retriever: ctx is nil")
	}

	deadline := ctx
	var cancel context.CancelFunc
	if r.maxTotalAttemptTime > 0 {
		deadline, cancel = context.WithTimeout(ctx, r.maxTotalAttemptTime)
		defer cancel()
	}

	var lastErr error
	for attempt := 0; attempt < r.maxAttempt; attempt++ {
		resp, retry, err := fn(deadline)
		lastErr = err

		if err == nil || !retry {
			return resp, err
		}

		if attempt == r.maxAttempt-1 {
			break
		}

		timer := time.NewTimer(r.backoff.Next(attempt))
		select {
		case <-deadline.Done():
			timer.Stop()
			return nil, deadline.Err()
		case <-timer.C:
		}
		timer.Stop()
	}

	return nil, fmt.Errorf("retriever: max attempts (%d) reached: %w", r.maxAttempt, lastErr)
}
