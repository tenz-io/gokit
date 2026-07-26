package async

import (
	"context"
	"errors"
	"sync"
)

// Option configures a [Group].
type Option[T any] func(*Group[T])

// WithLimit bounds the number of concurrent tasks to n. Tasks beyond the limit
// block until a slot frees. A non-positive limit is ignored (the default is
// unlimited). Has no effect on cancel-on-error behaviour.
func WithLimit[T any](n int) Option[T] {
	return func(g *Group[T]) {
		if n > 0 {
			g.limit = n
		}
	}
}

// WithCancelOnError makes [Group.Wait] cancel remaining tasks on the first
// failure (panic counts as a failure) and return that error. Without this
// option the group runs every task to completion and joins all errors with
// [errors.Join]. The derived context passed to tasks is cancelled on first
// error so in-flight tasks can short-circuit.
func WithCancelOnError[T any]() Option[T] {
	return func(g *Group[T]) {
		g.cancelOnError = true
	}
}

// Group is a generic, panic-safe errgroup-style builder for tasks that share a
// type parameter T. Add tasks with [Group.Go]; collect either an aggregated error
// ([Group.Wait]) or the ordered results ([Group.Results]). A zero Group is not
// usable — always obtain one via [New].
type Group[T any] struct {
	ctx           context.Context
	derived       context.Context
	cancel        context.CancelFunc
	limit         int
	cancelOnError bool

	sem      chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	errs     []error
	firstErr error // populated only when cancelOnError is set

	results []Result[T]
	started bool
}

// New returns a [Group] bound to ctx. Tasks added via [Group.Go] receive a
// derived context that is cancelled on first error only when
// [WithCancelOnError] is set; otherwise it simply tracks parent cancellation.
func New[T any](ctx context.Context, opts ...Option[T]) *Group[T] {
	derived, cancel := context.WithCancel(ctx)
	g := &Group[T]{
		ctx:     ctx,
		derived: derived,
		cancel:  cancel,
		results: nil, // grown on demand
	}
	for _, opt := range opts {
		opt(g)
	}
	if g.limit > 0 {
		g.sem = make(chan struct{}, g.limit)
	}
	return g
}

// Go submits a task for concurrent execution. It blocks until a concurrency
// slot is available when [WithLimit] is set. Calling Go concurrently with Wait
// is not supported: add all tasks first, then Wait.
//
// A nil task is a no-op. Results are collected in completion order; use
// [Group.Results] only after [Group.Wait] and only when not using
// [WithCancelOnError] (cancelled tasks produce no result).
func (g *Group[T]) Go(task Task[T]) {
	if task == nil {
		return
	}
	g.started = true
	g.wg.Add(1)
	if g.sem != nil {
		g.sem <- struct{}{}
	}
	go func() {
		defer g.wg.Done()
		if g.sem != nil {
			defer func() { <-g.sem }()
		}
		v, err := recoverTask(task)(g.derived)
		if err != nil {
			g.mu.Lock()
			// Under cancel-on-error, a task that fails because the group was
			// already cancelled (ctx.Err()) is a downstream symptom, not an
			// independent failure: drop it so Wait reports only the first error.
			if g.cancelOnError && g.firstErr != nil && errors.Is(err, context.Canceled) {
				g.mu.Unlock()
				return
			}
			g.errs = append(g.errs, err)
			if g.cancelOnError && g.firstErr == nil {
				g.firstErr = err
				g.cancel()
			}
			g.mu.Unlock()
			return
		}
		g.mu.Lock()
		g.results = append(g.results, Result[T]{Value: v})
		g.mu.Unlock()
	}()
}

// Wait blocks until all submitted tasks finish and returns the aggregated error.
// With [WithCancelOnError] it returns the first failure; otherwise it joins
// every error. A group with no tasks (or only nil tasks) returns nil.
func (g *Group[T]) Wait() error {
	g.wg.Wait()
	g.cancel()
	if g.cancelOnError && g.firstErr != nil {
		return g.firstErr
	}
	return errors.Join(g.errs...)
}

// Results returns the successful outcomes collected during [Group.Wait], in
// completion order. Failed tasks contribute nothing. Returns nil before Wait
// completes or when no task succeeded. The returned slice is a copy; mutate it
// freely without affecting the group.
func (g *Group[T]) Results() []Result[T] {
	if len(g.results) == 0 {
		return nil
	}
	out := make([]Result[T], len(g.results))
	copy(out, g.results)
	return out
}
