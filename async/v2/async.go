// Package async provides generic concurrent job execution with panic recovery.
package async

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
)

// Fn is a generic async function.
type Fn[T any] func(context.Context) (T, error)

// Holder holds a single result from AllOf, preserving input order.
type Holder[T any] struct {
	idx int
	Val T
	Err error
}

// Run runs all functions concurrently and returns the first error encountered.
// Uses errgroup semantics: the context is not cancelled on first error.
func Run[T any](ctx context.Context, fns ...Fn[T]) error {
	if len(fns) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(fns))

	for _, fn := range fns {
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func(f Fn[T]) {
			defer wg.Done()
			_, err := panicProof(f)(ctx)
			if err != nil {
				errCh <- err
			}
		}(fn)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		return err
	case <-done:
		return nil
	}
}

// Wait runs all functions concurrently and waits for all to complete, ignoring errors.
func Wait[T any](ctx context.Context, fns ...Fn[T]) {
	if len(fns) == 0 {
		return
	}
	var wg sync.WaitGroup
	for _, fn := range fns {
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func(f Fn[T]) {
			defer wg.Done()
			panicProof(f)(ctx) //nolint:errcheck
		}(fn)
	}
	wg.Wait()
}

// AllOf runs all functions concurrently and returns results in input order.
func AllOf[T any](ctx context.Context, fns []Fn[T]) []Holder[T] {
	if len(fns) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	results := make([]Holder[T], len(fns))

	for i, fn := range fns {
		if fn == nil {
			results[i] = Holder[T]{idx: i, Err: fmt.Errorf("nil function")}
			continue
		}
		wg.Add(1)
		go func(idx int, f Fn[T]) {
			defer wg.Done()
			val, err := panicProof(f)(ctx)
			results[idx] = Holder[T]{idx: idx, Val: val, Err: err}
		}(i, fn)
	}
	wg.Wait()
	return results
}

// AnyOf runs all functions concurrently and returns the first successful result.
// Cancels remaining goroutines via context cancellation on first success.
// If all fail, returns a combined error.
func AnyOf[T any](ctx context.Context, fns ...Fn[T]) (T, error) {
	var zero T
	if len(fns) == 0 {
		return zero, fmt.Errorf("async.AnyOf: empty function list")
	}

	allCtx, cancelAll := context.WithCancel(ctx)
	defer cancelAll()

	var wg sync.WaitGroup
	resultCh := make(chan T, 1)
	errCh := make(chan error, len(fns))

	for _, fn := range fns {
		if fn == nil {
			return zero, fmt.Errorf("async.AnyOf: nil function")
		}
		wg.Add(1)
		go func(f Fn[T]) {
			defer wg.Done()
			result, err := panicProof(f)(allCtx)
			if err != nil {
				errCh <- err
				return
			}
			select {
			case resultCh <- result:
				cancelAll()
			default:
			}
		}(fn)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errs []error
	for {
		select {
		case result := <-resultCh:
			return result, nil
		case err, ok := <-errCh:
			if !ok {
				return zero, fmt.Errorf("async.AnyOf: all %d jobs failed: %v", len(fns), errs)
			}
			errs = append(errs, err)
		}
	}
}

func panicProof[T any](fn Fn[T]) Fn[T] {
	return func(ctx context.Context) (result T, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("async: panic recovered: %v\n%s", rec, string(debug.Stack()))
				err = panicToError(rec)
			}
		}()
		return fn(ctx)
	}
}

func panicToError(rec any) error {
	switch r := rec.(type) {
	case string:
		return fmt.Errorf("panic: %s", r)
	case error:
		return fmt.Errorf("panic: %w", r)
	default:
		return fmt.Errorf("panic: %v", r)
	}
}
