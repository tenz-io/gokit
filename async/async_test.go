package async

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestAllOf(t *testing.T) {
	type args struct {
		ctx     context.Context
		jobList []RunnableJob
	}
	type after func(*testing.T, *args)
	tests := []struct {
		name            string
		args            args
		wantErrMsg      string
		wantErr         bool
		wantMaxDuration time.Duration
		after           after
	}{
		{
			name: "when all jobs are successful then return no error",
			args: args{
				ctx: context.Background(),
				jobList: []RunnableJob{
					NewJob(func(ctx context.Context) (int, error) {
						time.Sleep(10 * time.Millisecond)
						return 1, nil
					}, "job-1 failed"),
					NewJob(func(ctx context.Context) (string, error) {
						time.Sleep(15 * time.Millisecond)
						return "2", nil
					}, "job-2 failed"),
				},
			},
			wantErrMsg:      "",
			wantErr:         false,
			wantMaxDuration: 17 * time.Millisecond,
			after: func(t *testing.T, args *args) {
				if len(args.jobList) != 2 {
					t.Errorf("len(jobList) = %d, want 2", len(args.jobList))
					return
				}

				var (
					job1 = args.jobList[0].(*Job[int])
					job2 = args.jobList[1].(*Job[string])
				)

				if job1.Value() != 1 {
					t.Errorf("job1.Value() = %d, want 1", job1.Value())
					return
				}

				if job2.Value() != "2" {
					t.Errorf("job2.Value() = %s, want 2", job2.Value())
					return
				}

			},
		},
		{
			name: "when one job is failed then return error",
			args: args{
				ctx: context.Background(),
				jobList: []RunnableJob{
					NewJob(func(ctx context.Context) (int, error) {
						time.Sleep(10 * time.Millisecond)
						return 1, nil
					}, "job-1 failed"),
					NewJob(func(ctx context.Context) (*string, error) {
						time.Sleep(15 * time.Millisecond)
						return nil, errors.New("timeout")
					}, "job-2 failed"),
				},
			},
			wantErrMsg:      "job-2 failed",
			wantErr:         true,
			wantMaxDuration: 16 * time.Millisecond,
			after: func(t *testing.T, args *args) {
				if len(args.jobList) != 2 {
					t.Errorf("len(jobList) = %d, want 2", len(args.jobList))
					return
				}

				var (
					job1 = args.jobList[0].(*Job[int])
					job2 = args.jobList[1].(*Job[*string])
				)

				if job1.Value() != 1 {
					t.Errorf("job1.Value() = %d, want 1", job1.Value())
					return
				}

				if job2.Value() != nil {
					t.Errorf("job2.Value() = %v, want nil", job2.Value())
					return
				}

			},
		},
		{
			name: "when all jobs are failed then return error",
			args: args{
				ctx: context.Background(),
				jobList: []RunnableJob{
					NewJob(func(ctx context.Context) (int, error) {
						time.Sleep(10 * time.Millisecond)
						return 0, errors.New("timeout")
					}, "job-1 failed"),
					NewJob(func(ctx context.Context) (*string, error) {
						time.Sleep(15 * time.Millisecond)
						return nil, errors.New("timeout")
					}, "job-2 failed"),
				},
			},
			wantErrMsg:      "job-1 failed",
			wantErr:         true,
			wantMaxDuration: 17 * time.Millisecond,
			after: func(t *testing.T, args *args) {
				if len(args.jobList) != 2 {
					t.Errorf("len(jobList) = %d, want 2", len(args.jobList))
					return
				}

				var (
					job1 = args.jobList[0].(*Job[int])
					job2 = args.jobList[1].(*Job[*string])
				)

				if job1.Value() != 0 {
					t.Errorf("job1.Value() = %d, want 1", job1.Value())
					return
				}

				if job2.Value() != nil {
					t.Errorf("job2.Value() = %v, want nil", job2.Value())
					return
				}

			},
		},
		{
			name: "when one job panic then return error",
			args: args{
				ctx: context.Background(),
				jobList: []RunnableJob{
					NewJob(func(ctx context.Context) (int, error) {
						time.Sleep(10 * time.Millisecond)
						panic("panic")
					}, "job-1 failed"),
					NewJob(func(ctx context.Context) (*string, error) {
						time.Sleep(15 * time.Millisecond)
						return nil, errors.New("timeout")
					}, "job-2 failed"),
				},
			},
			wantErrMsg:      "job-1 failed",
			wantErr:         true,
			wantMaxDuration: 17 * time.Millisecond,
			after: func(t *testing.T, args *args) {
				if len(args.jobList) != 2 {
					t.Errorf("len(jobList) = %d, want 2", len(args.jobList))
					return
				}

				var (
					job1 = args.jobList[0].(*Job[int])
					job2 = args.jobList[1].(*Job[*string])
				)

				if job1.Value() != 0 {
					t.Errorf("job1.Value() = %d, want 1", job1.Value())
					return
				}

				if job2.Value() != nil {
					t.Errorf("job2.Value() = %v, want nil", job2.Value())
					return
				}

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			gotErrMsg, err := AllOf(tt.args.ctx, tt.args.jobList...)
			duration := time.Since(start)
			t.Logf("duration: %s, gotErrMsg: %s, err: %v", duration, gotErrMsg, err)

			if (err != nil) != tt.wantErr {
				t.Errorf("Submit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotErrMsg != tt.wantErrMsg {
				t.Errorf("Submit() gotErrMsg = %v, want %v", gotErrMsg, tt.wantErrMsg)
			}

			if duration > tt.wantMaxDuration {
				t.Errorf("Submit() duration = %v, want %v", duration, tt.wantMaxDuration)
			}

			if tt.after != nil {
				tt.after(t, &tt.args)
			}

		})
	}
}

func TestOneOf(t *testing.T) {
	type args[T any] struct {
		ctx    context.Context
		fnList []Fn[string]
	}
	type after func(*testing.T, *args[string])
	type testCase[T any] struct {
		name            string
		args            args[T]
		want            T
		wantErr         bool
		wantMaxDuration time.Duration
		after           after
	}
	tests := []testCase[string]{
		{
			name: "when all jobs are successful then return the fastest job result",
			args: args[string]{
				ctx: context.Background(),
				fnList: []Fn[string]{
					func(ctx context.Context) (string, error) {
						time.Sleep(10 * time.Millisecond)
						return "1", nil
					},
					func(ctx context.Context) (string, error) {
						time.Sleep(5 * time.Millisecond)
						return "2", nil
					},
				},
			},
			want:            "2",
			wantErr:         false,
			wantMaxDuration: 7 * time.Millisecond,
		},
		{
			name: "when one job is failed and one job is successful then return the successful job result",
			args: args[string]{
				ctx: context.Background(),
				fnList: []Fn[string]{
					func(ctx context.Context) (string, error) {
						time.Sleep(10 * time.Millisecond)
						return "1", nil
					},
					func(ctx context.Context) (string, error) {
						time.Sleep(5 * time.Millisecond)
						return "", errors.New("oops")
					},
				},
			},
			want:            "1",
			wantErr:         false,
			wantMaxDuration: 12 * time.Millisecond,
		},
		{
			name: "when all jobs are failed then return error",
			args: args[string]{
				ctx: context.Background(),
				fnList: []Fn[string]{
					func(ctx context.Context) (string, error) {
						time.Sleep(10 * time.Millisecond)
						return "", errors.New("oops1")
					},
					func(ctx context.Context) (string, error) {
						time.Sleep(5 * time.Millisecond)
						return "", errors.New("oops2")
					},
				},
			},
			want:            "",
			wantErr:         true,
			wantMaxDuration: 12 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := OneOf(tt.args.ctx, tt.args.fnList...)
			duration := time.Since(start)
			t.Logf("duration: %s, got: %s, err: %v", duration, got, err)

			if (err != nil) != tt.wantErr {
				t.Errorf("OneOf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OneOf() got = %v, want %v", got, tt.want)
			}

			if duration > tt.wantMaxDuration {
				t.Errorf("OneOf() duration = %v, want %v", duration, tt.wantMaxDuration)
				return
			}

			if tt.after != nil {
				tt.after(t, &tt.args)
			}

		})
	}
}
