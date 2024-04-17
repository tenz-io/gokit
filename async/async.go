package async

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Fn[T any] func(context.Context) (T, error)

type RunnableJob interface {
	run(ctx context.Context) error
	errorMessage() string
	isNil() bool
}

type Job[T any] struct {
	fn  Fn[T]
	val T

	errMsg string
}

func NewJob[T any](fn Fn[T], errMsg string) *Job[T] {
	job := &Job[T]{
		errMsg: errMsg,
		fn:     fn,
	}

	return job
}

func (j *Job[T]) run(ctx context.Context) error {
	result, err := j.fn(ctx)
	if err != nil {
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

func (j *Job[T]) ValueOrZero() T {
	if j == nil {
		var zeroV T
		return zeroV
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

func errorFromPanic(i any) error {
	switch err := i.(type) {
	case string:
		return fmt.Errorf(err)
	case error:
		return err
	default:
		return fmt.Errorf("unknown panic")

	}
}

type Builder struct {
	jobList []RunnableJob
}

func NewBuilder() *Builder {
	return &Builder{
		jobList: make([]RunnableJob, 0),
	}
}

func (p *Builder) AddJob(jobs ...RunnableJob) {
	p.jobList = append(p.jobList, jobs...)
}

// Run runs all jobs concurrently and returns the first error encountered.
func (p *Builder) Run(ctx context.Context) (errMsg string, err error) {
	return AllOf(ctx, p.jobList...)
}

// AllOf runs all jobs concurrently and returns the first error encountered.
func AllOf(ctx context.Context, jobList ...RunnableJob) (errMsg string, err error) {
	wge, wgCtx := errgroup.WithContext(ctx)

	for _, job := range jobList {
		if job == nil || job.isNil() {
			continue
		}

		tempJob := job

		wge.Go(func() (innerErr error) {
			defer func() {
				if rec := recover(); rec != nil {
					innerErr = &errPair{
						errMsg: tempJob.errorMessage(),
						err:    errorFromPanic(rec),
					}
				}
			}()

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

// OneOf runs all jobs concurrently and returns the first job result that is not error.
func OneOf[T any](ctx context.Context, fnList ...Fn[T]) (T, error) {
	resultC := make(chan T, len(fnList))
	errC := make(chan error, len(fnList))
	for _, fn := range fnList {
		if fn == nil {
			var zero T
			return zero, fmt.Errorf("has nil function")
		}

		newCtx, cancel := context.WithCancel(ctx)
		go func(f Fn[T]) {
			defer cancel()
			result, err := f(newCtx)
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
				var zero T
				return zero, fmt.Errorf("all jobs are failed, one of error: %w", err)
			}
		}
	}
}
