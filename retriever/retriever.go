package retriever

import (
	"context"
	"fmt"
	"time"
)

// HealthCheckFunc specifies the behaviour for custom health check.
// If the health check returns error then retriever fails.
//type HealthCheckFunc func() error

// DoFunc is the function which retriever will call on loop until one of this condition is met:
// - maxAttempt has been reached
// - retry return parameter is false
// - context deadline
// - err return parameter is nil
type DoFunc func(context.Context) (resp any, retry bool, err error)

// DoFuncAlwaysRetry is the function which retriever will call on loop until one of this condition is met:
// - maxAttempt has been reached
// - context deadline
// - err return parameter is nil
type DoFuncAlwaysRetry func(context.Context) (resp any, err error)

type Retriever interface {
	// DoAlwaysRetry will call doFunc until it returns nil or context deadline
	// If doFunc returns error, it will retry until maxAttempt is reached
	DoAlwaysRetry(ctx context.Context, doFunc DoFuncAlwaysRetry) (any, error)
	// Do will call doFunc until it returns nil or context deadline
	// If doFunc returns error, it will retry until maxAttempt is reached
	Do(ctx context.Context, doFunc DoFunc) (any, error)
}

type retriever struct {
	maxAttempt          int
	maxTotalAttemptTime time.Duration
	backoff             Backoff
	useGoroutine        bool
}

func NewRetrieverWithConfig(configOpts ...ConfigOpt) Retriever {
	config := defaultConfig
	for _, configOpt := range configOpts {
		configOpt(&config)
	}

	return NewRetriever(config)
}

func NewRetriever(config Config) Retriever {
	if config.MaxAttempt <= 0 {
		config.MaxAttempt = 3
	}

	if config.Backoff == nil {
		// ExponentialBackoff with 100ms base and 2.0 factor
		config.Backoff = NewExponentialBackoff(100, 2.0, 0.3)
	}

	return &retriever{
		maxAttempt:          config.MaxAttempt,
		maxTotalAttemptTime: config.MaxTotalAttemptTime,
		backoff:             config.Backoff,
		useGoroutine:        config.UseGoroutine,
	}
}

func (r *retriever) DoAlwaysRetry(ctx context.Context, doFunc DoFuncAlwaysRetry) (any, error) {
	return r.Do(ctx, func(funcCtx context.Context) (any, bool, error) {
		output, err := doFunc(funcCtx)
		return output, true, err
	})
}

func (r *retriever) Do(ctx context.Context, doFunc DoFunc) (any, error) {
	if doFunc == nil {
		return nil, fmt.Errorf("doFunc cannot be nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("ctx cannot be nil")
	}

	var (
		newCtx context.Context
		cancel context.CancelFunc
	)

	if r.maxTotalAttemptTime > 0 {
		newCtx, cancel = context.WithTimeout(ctx, r.maxTotalAttemptTime)
		defer cancel()
	} else {
		newCtx = ctx
	}

	var lastError error

	for failCount := 0; failCount < r.maxAttempt; failCount++ {
		var (
			resp      any
			retryable bool
			err       error
		)

		if r.useGoroutine {
			doneC := make(chan any)
			go func() {
				defer close(doneC)
				resp, retryable, err = doFunc(newCtx)
			}()

			select {
			case <-newCtx.Done():
				return nil, newCtx.Err()
			case <-doneC:
			}
		} else {
			resp, retryable, err = doFunc(newCtx)
		}

		lastError = err

		// Success or non retryable error
		if err == nil || !retryable {
			return resp, err
		}

		retryWait := time.NewTimer(r.backoff.Next(failCount))

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
