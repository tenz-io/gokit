// Package async provides generic concurrent task execution with panic recovery.
//
// The package offers two complementary styles:
//
//   - Stateless helpers [Run], [RunAll], [Wait], [AllOf], [AnyOf] for one-shot
//     fan-out where the full task set is known up front.
//   - A [Group] builder (see [New]) for open-ended use where tasks are added
//     incrementally, with optional concurrency limit and cancel-on-first-error.
//
// Every task runs inside a panic-safe wrapper that converts a panic into a
// typed [*PanicError], so a panicking task can never crash the process.
package async

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
)

// Task is the shared async task signature. All helpers and [Group] operate on it.
type Task[T any] func(context.Context) (T, error)

// Result holds the outcome of a single task. Slice position encodes input order,
// so there is no separate index field.
type Result[T any] struct {
	Value T
	Err   error
}

// PanicError is returned when a task panics. It carries the recovered value and a
// captured stack trace so callers can inspect or log it via [errors.As].
type PanicError struct {
	value any
	stack []byte
}

// Value returns the value passed to panic.
func (p *PanicError) Value() any { return p.value }

// Stack returns the goroutine stack captured at the panic site.
func (p *PanicError) Stack() []byte { return p.stack }

// Error implements the error interface.
func (p *PanicError) Error() string {
	switch v := p.value.(type) {
	case error:
		return fmt.Sprintf("async: panic: %v", v)
	case string:
		return fmt.Sprintf("async: panic: %s", v)
	default:
		return fmt.Sprintf("async: panic: %v", v)
	}
}

// Unwrap allows errors.Is/As to reach into panics that wrapped an error.
func (p *PanicError) Unwrap() error {
	if err, ok := p.value.(error); ok {
		return err
	}
	return nil
}

// Run executes the tasks concurrently and returns the first error encountered,
// nil-filtered. It does not cancel sibling tasks on error (no context is derived).
// Panics are reported as [*PanicError]. Returns nil for an empty task set.
func Run[T any](ctx context.Context, tasks ...Task[T]) error {
	tasks = filterNil(tasks)
	if len(tasks) == 0 {
		return nil
	}

	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		first error
	)

	for _, task := range tasks {
		wg.Add(1)
		go func(f Task[T]) {
			defer wg.Done()
			if _, err := recoverTask(f)(ctx); err != nil {
				mu.Lock()
				if first == nil {
					first = err
				}
				mu.Unlock()
			}
		}(task)
	}
	wg.Wait()
	return first
}

// RunAll executes the tasks concurrently and joins every non-nil error with
// [errors.Join]. Panics are reported as [*PanicError]. Returns nil when every
// task succeeds or the task set is empty.
func RunAll[T any](ctx context.Context, tasks ...Task[T]) error {
	tasks = filterNil(tasks)
	if len(tasks) == 0 {
		return nil
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, task := range tasks {
		wg.Add(1)
		go func(f Task[T]) {
			defer wg.Done()
			if _, err := recoverTask(f)(ctx); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(task)
	}
	wg.Wait()
	return errors.Join(errs...)
}

// Wait executes the tasks concurrently and blocks until all finish, discarding
// both values and errors. It is the fire-and-forget variant. Panics are swallowed
// (recovered but not returned); use [Run] or [RunAll] if they must be observed.
func Wait[T any](ctx context.Context, tasks ...Task[T]) {
	tasks = filterNil(tasks)
	if len(tasks) == 0 {
		return
	}
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(f Task[T]) {
			defer wg.Done()
			_, _ = recoverTask(f)(ctx)
		}(task)
	}
	wg.Wait()
}

// AllOf executes the tasks concurrently and returns one [Result] per task, in
// input order. A task failing or panicking does not affect the others; its
// result carries the error (panic → [*PanicError]). Returns nil for an empty set.
func AllOf[T any](ctx context.Context, tasks ...Task[T]) []Result[T] {
	tasks = filterNil(tasks)
	if len(tasks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	results := make([]Result[T], len(tasks))
	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, f Task[T]) {
			defer wg.Done()
			v, err := recoverTask(f)(ctx)
			results[idx] = Result[T]{Value: v, Err: err}
		}(i, task)
	}
	wg.Wait()
	return results
}

// AnyOf executes the tasks concurrently and returns the first successful result,
// cancelling the remaining tasks through a derived context on first success.
// If every task fails, the failures are joined with [errors.Join]. Panics are
// treated as failures ([*PanicError]). An empty task set is an error.
func AnyOf[T any](ctx context.Context, tasks ...Task[T]) (T, error) {
	var zero T
	if len(tasks) == 0 {
		return zero, errors.New("async.AnyOf: empty task set")
	}
	for _, t := range tasks {
		if t == nil {
			return zero, errors.New("async.AnyOf: nil task")
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errs   []error
		won    bool
		winner T
	)

	for _, task := range tasks {
		wg.Add(1)
		go func(f Task[T]) {
			defer wg.Done()
			v, err := recoverTask(f)(ctx)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			if !won {
				won = true
				winner = v
				cancel()
			}
			mu.Unlock()
		}(task)
	}

	wg.Wait()
	if won {
		return winner, nil
	}
	return zero, fmt.Errorf("async.AnyOf: all %d task(s) failed: %w", len(tasks), errors.Join(errs...))
}

// recoverTask wraps a task so that a panic is recovered and surfaced as a
// [*PanicError] instead of crashing the process. The stack is captured once,
// at the panic site.
func recoverTask[T any](task Task[T]) Task[T] {
	return func(ctx context.Context) (result T, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				err = &PanicError{value: rec, stack: stack()}
			}
		}()
		return task(ctx)
	}
}

// stack returns the current goroutine stack, reused by [PanicError].
func stack() []byte {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return buf[:n]
}

// filterNil returns tasks with nil entries dropped, preserving order.
func filterNil[T any](tasks []Task[T]) []Task[T] {
	out := tasks[:0:0]
	for _, t := range tasks {
		if t != nil {
			out = append(out, t)
		}
	}
	return out
}
