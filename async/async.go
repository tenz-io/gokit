package async

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Holder[T any] struct {
	idx int
	Val T
	Err error
}

type Fn[T any] func(context.Context) (T, error)

// job is an interface that represents a job that can be run concurrently.
// make it private to prevent other package to implement it.
type job interface {
	run(ctx context.Context) error
	errorMessage() string
	isNil() bool
}

func NewJob[T any](fn Fn[T], errMsg string) *Job[T] {
	return &Job[T]{
		errMsg: errMsg,
		fn:     fn,
	}
}

type Job[T any] struct {
	fn  Fn[T]
	err error
	val T

	errMsg string
}

func (j *Job[T]) run(ctx context.Context) error {
	result, err := withPanicProof(j.fn)(ctx)
	if err != nil {
		j.err = err
		return err
	}

	j.val = result
	return nil
}

func (j *Job[T]) errorMessage() string {
	return j.errMsg
}

func (j *Job[T]) Value() T {
	return j.val
}

func (j *Job[T]) isNil() bool {
	return j == nil
}

func (j *Job[T]) Result() (T, error) {
	return j.val, j.err
}

func (j *Job[T]) ValueOrZero() T {
	if j == nil {
		var zero T
		return zero
	}

	return j.val
}

type errPair struct {
	errMsg string
	err    error
}

func (e *errPair) Error() string {
	return e.err.Error()
}

type Builder struct {
	jobList []job
}

func NewBuilder() *Builder {
	return &Builder{
		jobList: make([]job, 0),
	}
}

func (p *Builder) AddJob(jobs ...job) {
	p.jobList = append(p.jobList, jobs...)
}

// Run runs all jobs concurrently and returns the first error encountered.
func (p *Builder) Run(ctx context.Context) (errMsg string, err error) {
	return AllOf(ctx, p.jobList...)
}

// Submit runs all jobs concurrently and returns when all jobs are done.
func (p *Builder) Submit(ctx context.Context) {
	Submit(ctx, p.jobList...)
}

// Submit runs all jobs concurrently and returns when all jobs are done.
func Submit(ctx context.Context, jobs ...job) {
	if len(jobs) == 0 {
		return
	}

	wg := sync.WaitGroup{}
	for _, jb := range jobs {
		if jb == nil || jb.isNil() {
			continue
		}

		newCtx := context.WithoutCancel(ctx)
		wg.Add(1)
		go func(rj job) {
			defer wg.Done()
			_ = rj.run(newCtx)
		}(jb)
	}

	wg.Wait()

}

// AllOf runs all jobs concurrently and returns the first error encountered.
func AllOf(ctx context.Context, jobs ...job) (errMsg string, err error) {
	if len(jobs) == 0 {
		return "", fmt.Errorf("empty job list")
	}

	wge, wgCtx := errgroup.WithContext(ctx)
	for _, jb := range jobs {
		if jb == nil || jb.isNil() {
			continue
		}

		tempJob := jb
		wge.Go(func() (innerErr error) {
			if jobErr := tempJob.run(wgCtx); jobErr != nil {
				return &errPair{
					errMsg: tempJob.errorMessage(),
					err:    jobErr,
				}
			}

			return nil
		})
	}

	if groupErr := wge.Wait(); groupErr != nil {
		var pair *errPair
		if ok := errors.As(groupErr, &pair); !ok {
			return "", groupErr
		}

		return pair.errMsg, pair.err
	}

	return "", nil
}

// AnyOf runs all jobs concurrently and returns the fastest job result that is not error.
func AnyOf[T any](ctx context.Context, fnList ...Fn[T]) (T, error) {
	var (
		zero T
	)

	if len(fnList) == 0 {
		return zero, fmt.Errorf("empty function list")
	}

	allCtx, cancelAll := context.WithCancel(ctx)
	defer cancelAll()

	resultC := make(chan T, len(fnList))
	errC := make(chan error, len(fnList))
	for _, fn := range fnList {
		if fn == nil {
			return zero, fmt.Errorf("has nil function")
		}

		newCtx := context.WithoutCancel(allCtx)
		go func(f Fn[T]) {
			result, err := withPanicProof(f)(newCtx)
			if err != nil {
				errC <- err
				return
			}
			resultC <- result
		}(fn)
	}

	errCount := 0
	for {
		select {
		case result := <-resultC:
			return result, nil
		case err := <-errC:
			errCount++
			if errCount == len(fnList) {
				return zero, fmt.Errorf("all jobs are failed, one of error: %w", err)
			}
		}
	}
}

// withPanicProof is a wrapper function to catch panic and convert it to error.
// It is useful to prevent the application from crashing due to panic.
// The panic message and stack trace will be logged.
func withPanicProof[T any](fn Fn[T]) Fn[T] {
	return func(ctx context.Context) (result T, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovery: %s, stacktrace: %s\n", rec, string(debug.Stack()))
				err = errorFromPanic(rec)
			}
		}()

		result, err = fn(ctx)
		return
	}
}

// errorFromPanic converts panic to error.
func errorFromPanic(rec any) error {
	switch rt := rec.(type) {
	case string:
		return fmt.Errorf(rt)
	case error:
		return rt
	default:
		return fmt.Errorf("unknown panic")
	}
}

// Run runs all jobs concurrently and returns the results.
// The order of the results is the same as the order of the input functions.
func Run[T any](ctx context.Context, fnList []Fn[T]) (results []Holder[T]) {
	var (
		count = len(fnList)
	)

	if count == 0 {
		return []Holder[T]{}
	}

	resultC := make(chan *Holder[T], count)
	for idx, fn := range fnList {
		if fn == nil {
			var zero T
			resultC <- &Holder[T]{
				idx: idx,
				Val: zero,
				Err: fmt.Errorf("nil function"),
			}
			continue
		}

		newCtx := context.WithoutCancel(ctx)
		go func(i int, f Fn[T]) {
			result, err := withPanicProof(f)(newCtx)
			resultC <- &Holder[T]{
				idx: i,
				Val: result,
				Err: err,
			}
		}(idx, fn)
	}

	results = make([]Holder[T], count)
	for i := 0; i < count; i++ {
		select {
		case result := <-resultC:
			results[result.idx] = *result
		}
	}

	return results
}
