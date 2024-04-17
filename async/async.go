package async

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type RunnableJob interface {
	run(ctx context.Context) error
	errorMessage() string
	isNil() bool
}

type Job[Val any] struct {
	fn  func(ctx context.Context) (Val, error)
	val Val

	errMsg string
}

func NewJob[Val any](fn func(ctx context.Context) (Val, error), errMsg string) *Job[Val] {
	job := &Job[Val]{
		errMsg: errMsg,
		fn:     fn,
	}

	return job
}

func (j *Job[Val]) run(ctx context.Context) error {
	result, err := j.fn(ctx)
	if err != nil {
		return err
	}

	j.val = result
	return nil
}

func (j *Job[Val]) errorMessage() string {
	return j.errMsg
}

func (j *Job[Val]) Value() Val {
	return j.val
}

func (j *Job[Val]) isNil() bool {
	return j == nil
}

func (j *Job[Val]) ValueOrZero() Val {
	if j == nil {
		var zeroV Val
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

func (p *Builder) Run(ctx context.Context) (errMsg string, err error) {
	return Submit(ctx, p.jobList...)
}

func Submit(ctx context.Context, jobList ...RunnableJob) (errMsg string, err error) {
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
