package retriever

import (
	"context"
	"fmt"
	"time"
)

// HealthCheckFunc specifies the behaviour for custom health check.
// If the health check returns error then retriever fails.
type HealthCheckFunc func() error

// DoFunc is the function which retriever will call on loop until one of this condition is met
// - maxAttempt has been reached
// - retry return parameter is false
// - context deadline
// - err return parameter is nil
type DoFunc func(context.Context) (resp any, retry bool, err error)

// DoFuncAlwaysRetry is the function which retriever will call on loop until one of this condition is met
// - maxAttempt has been reached
// - context deadline
// - err return parameter is nil
type DoFuncAlwaysRetry func(context.Context) (resp any, err error)

// Config for Retriever, all fields are optional
type Config struct {
	MaxAttempt          int           // Maximum number of attempt before failing (default: 3)
	MaxTotalAttemptTime time.Duration // Maximum total duration to wait for result (default: unlimited/until context expired)

	// An implementation of Backoff to calculate the time duration to wait for the next operation.
	// Will default to ExponentialBackoff with 100ms base and 2.0 factor
	Backoff Backoff

	// If true, DoFunc will be executed using goroutine. This allows for long function without context support (default: false)
	UseGoroutine bool
}

type Retriever struct {
	maxAttempt          int
	maxTotalAttemptTime time.Duration

	backoff Backoff

	useGoroutine bool
}

func NewRetriever(config Config) *Retriever {
	if config.MaxAttempt <= 0 {
		config.MaxAttempt = 3
	}

	if config.Backoff == nil {
		// ExponentialBackoff with 100ms base and 2.0 factor
		config.Backoff = NewExponentialBackoff(100, 2.0, 0.3)
	}

	return &Retriever{
		maxAttempt:          config.MaxAttempt,
		maxTotalAttemptTime: config.MaxTotalAttemptTime,
		backoff:             config.Backoff,
		useGoroutine:        config.UseGoroutine,
	}
}

func (r *Retriever) DoAlwaysRetry(ctx context.Context, doFunc DoFuncAlwaysRetry) (any, error) {
	return r.Do(ctx, func(funcCtx context.Context) (any, bool, error) {
		output, err := doFunc(funcCtx)
		return output, true, err
	})
}

func (r *Retriever) Do(ctx context.Context, doFunc DoFunc) (any, error) {
	if doFunc == nil {
		return nil, fmt.Errorf("doFunc cannot be nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("ctx cannot be nil")
	}

	var newCtx context.Context
	var cancel context.CancelFunc

	if r.maxTotalAttemptTime != 0 {
		newCtx, cancel = context.WithTimeout(ctx, r.maxTotalAttemptTime)
		defer cancel()
	} else {
		newCtx = ctx
	}

	var lastError error

	for failureCount := 0; failureCount < r.maxAttempt; failureCount++ {
		var resp any
		var retryable bool
		var err error

		if r.useGoroutine {
			doneCh := make(chan any)
			go func() {
				resp, retryable, err = doFunc(newCtx)
				close(doneCh)
			}()

			select {
			case <-newCtx.Done():
				return nil, newCtx.Err()
			case <-doneCh:
			}
		} else {
			resp, retryable, err = doFunc(newCtx)
		}

		lastError = err

		// Success or non retryable error
		if err == nil || !retryable {
			return resp, err
		}

		retryWait := time.NewTimer(r.backoff.Next(failureCount))

		select {
		case <-newCtx.Done():
			retryWait.Stop()
			return nil, newCtx.Err()
		case <-retryWait.C:
		}

		retryWait.Stop()
	}

	return nil, fmt.Errorf("max retry reached, err: %w", lastError)
}
