package retriever

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
	"time"
)

type MockObject struct {
	count int
	mock.Mock
}

func (m *MockObject) MockDoFunc(ctx context.Context) (any, bool, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Bool(1), args.Error(2)
}

func (m *MockObject) MockDoAlwaysFunc(ctx context.Context) (any, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

func Test_retriever_Do(t *testing.T) {
	type fields struct {
		maxAttempt          int
		maxTotalAttemptTime time.Duration
		backoff             Backoff
		useGoroutine        bool
	}
	type args struct {
		ctx context.Context
	}
	type behavior func(*MockObject)
	tests := []struct {
		name                string
		fields              fields
		args                args
		behavior            behavior
		want                any
		wantErr             assert.ErrorAssertionFunc
		wantDoFuncNumCalled int
		wantMaxDuration     time.Duration
	}{
		{
			name: "when ctx is nil then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: nil,
			},
			behavior: func(mockedObj *MockObject) {
				// do nothing
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 0,
			wantMaxDuration:     5 * time.Millisecond,
		},
		{
			name: "when doFunc returns error then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoFunc", mock.Anything).After(10*time.Millisecond).
					Return(nil, false, assert.AnError)
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 1,
			wantMaxDuration:     15 * time.Millisecond,
		},
		{
			name: "when doFunc returns error with retry then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoFunc", mock.Anything).After(10*time.Millisecond).
					Return(nil, true, assert.AnError)
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 3,
			wantMaxDuration:     35 * time.Millisecond,
		},
		{
			name: "when doFunc returns output then return output",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoFunc", mock.Anything).After(10*time.Millisecond).
					Return("output", false, nil)
			},
			want:                "output",
			wantErr:             assert.NoError,
			wantDoFuncNumCalled: 1,
			wantMaxDuration:     15 * time.Millisecond,
		},
		{
			name: "when doFunc retry returns output then return output",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				// first call
				mockedObj.On("MockDoFunc", mock.Anything).Times(1).After(10*time.Millisecond).
					Return("", true, assert.AnError)
				// retry
				mockedObj.On("MockDoFunc", mock.Anything).Times(1).After(20*time.Millisecond).
					Return("output", true, nil)
			},
			want:                "output",
			wantErr:             assert.NoError,
			wantDoFuncNumCalled: 2,
			wantMaxDuration:     35 * time.Millisecond,
		},
		{
			name: "when doFunc retry again returns output then return output",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				// the first two calls return error
				mockedObj.On("MockDoFunc", mock.Anything).Times(2).After(10*time.Millisecond).
					Return("", true, assert.AnError)
				// the third call return output
				mockedObj.On("MockDoFunc", mock.Anything).Times(1).After(10*time.Millisecond).
					Return("output", true, nil)
			},
			want:                "output",
			wantErr:             assert.NoError,
			wantDoFuncNumCalled: 3,
			wantMaxDuration:     35 * time.Millisecond,
		},
		{
			name: "when doFunc returns execution time more than maxTotalAttemptTime then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        true,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoFunc", mock.Anything).After(120*time.Millisecond).
					Return("output", true, nil)
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 1,
			wantMaxDuration:     110 * time.Millisecond,
		},
		{
			name: "when doFunc returns total execution time more than maxTotalAttemptTime then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        true,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoFunc", mock.Anything).After(40*time.Millisecond).
					Return(nil, true, assert.AnError)
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 3,
			wantMaxDuration:     110 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &retriever{
				maxAttempt:          tt.fields.maxAttempt,
				maxTotalAttemptTime: tt.fields.maxTotalAttemptTime,
				backoff:             tt.fields.backoff,
				useGoroutine:        tt.fields.useGoroutine,
			}

			mockedObj := new(MockObject)
			tt.behavior(mockedObj)

			startTime := time.Now()
			got, err := r.Do(tt.args.ctx, mockedObj.MockDoFunc)
			duration := time.Since(startTime)
			t.Logf("Do() got = %v, err = %+v, duration: %s", got, err, duration)

			if ret := tt.wantErr(t, err, "Do() error = %+v, wantErr %v", err, true); !ret {
				t.Errorf("Do() error = %+v, wantErr %v", err, ret)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Do() got = %v, want %v", got, tt.want)
			}

			mockedObj.AssertNumberOfCalls(t, "MockDoFunc", tt.wantDoFuncNumCalled)

			if tt.wantMaxDuration > 0 && duration > tt.wantMaxDuration {
				t.Errorf("Do() duration = %v, want less than %v", duration, tt.wantMaxDuration)
				return
			}

		})
	}
}

func Test_retriever_DoAlwaysRetry(t *testing.T) {
	type fields struct {
		maxAttempt          int
		maxTotalAttemptTime time.Duration
		backoff             Backoff
		useGoroutine        bool
	}
	type args struct {
		ctx context.Context
	}
	type behavior func(*MockObject)
	tests := []struct {
		name                string
		fields              fields
		args                args
		behavior            behavior
		want                any
		wantErr             assert.ErrorAssertionFunc
		wantDoFuncNumCalled int
		wantMaxDuration     time.Duration
	}{
		{
			name: "when ctx is nil then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: nil,
			},
			behavior: func(mockedObj *MockObject) {
				// do nothing
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 0,
			wantMaxDuration:     5 * time.Millisecond,
		},
		{
			name: "when doFunc returns error then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).After(10*time.Millisecond).
					Return(nil, assert.AnError)
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 3,
			wantMaxDuration:     35 * time.Millisecond,
		},
		{
			name: "when doFunc returns output then return output",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).After(10*time.Millisecond).
					Return("output", nil)
			},
			want:                "output",
			wantErr:             assert.NoError,
			wantDoFuncNumCalled: 1,
			wantMaxDuration:     15 * time.Millisecond,
		},
		{
			name: "when doFunc retry returns output then return output",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				// first call
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).Times(1).After(10*time.Millisecond).
					Return("", assert.AnError)
				// retry
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).Times(1).After(20*time.Millisecond).
					Return("output", nil)
			},
			want:                "output",
			wantErr:             assert.NoError,
			wantDoFuncNumCalled: 2,
			wantMaxDuration:     35 * time.Millisecond,
		},
		{
			name: "when doFunc retry again returns output then return output",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        false,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				// the first two calls return error
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).Times(2).After(10*time.Millisecond).
					Return("", assert.AnError)
				// the third call return output
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).Times(1).After(10*time.Millisecond).
					Return("output", nil)
			},
			want:                "output",
			wantErr:             assert.NoError,
			wantDoFuncNumCalled: 3,
			wantMaxDuration:     35 * time.Millisecond,
		},
		{
			name: "when doFunc returns execution time more than maxTotalAttemptTime then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        true,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).After(120*time.Millisecond).
					Return("output", nil)
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 1,
			wantMaxDuration:     110 * time.Millisecond,
		},
		{
			name: "when doFunc returns total execution time more than maxTotalAttemptTime then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				backoff:             &NoBackoff{},
				useGoroutine:        true,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).After(40*time.Millisecond).
					Return(nil, assert.AnError)
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 3,
			wantMaxDuration:     110 * time.Millisecond,
		},
		{
			name: "when doFunc exponential backoff returns total execution time more than maxTotalAttemptTime then return error",
			fields: fields{
				maxAttempt:          3,
				maxTotalAttemptTime: 100 * time.Millisecond,
				// wait:  ~20ms, ~40ms, ~80ms
				backoff:      NewExponentialBackoff(20, 2.0, 0.3),
				useGoroutine: true,
			},
			args: args{
				ctx: context.Background(),
			},
			behavior: func(mockedObj *MockObject) {
				// exe:25ms + wait:20ms + exe:25ms + wait:40ms => 110ms timeout, so only 2 times can be executed
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).Times(2).After(25*time.Millisecond).
					Return(nil, assert.AnError)
				mockedObj.On("MockDoAlwaysFunc", mock.Anything).Times(1).After(20*time.Millisecond).
					Return("output", nil)
			},
			want:                nil,
			wantErr:             assert.Error,
			wantDoFuncNumCalled: 2,
			wantMaxDuration:     110 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &retriever{
				maxAttempt:          tt.fields.maxAttempt,
				maxTotalAttemptTime: tt.fields.maxTotalAttemptTime,
				backoff:             tt.fields.backoff,
				useGoroutine:        tt.fields.useGoroutine,
			}

			mockedObj := new(MockObject)
			tt.behavior(mockedObj)

			startTime := time.Now()
			got, err := r.DoAlwaysRetry(tt.args.ctx, mockedObj.MockDoAlwaysFunc)
			duration := time.Since(startTime)
			t.Logf("DoAlwaysRetry() got = %v, err = %+v, duration: %s", got, err, duration)

			if ret := tt.wantErr(t, err, "Do() error = %+v, wantErr %v", err, true); !ret {
				t.Errorf("DoAlwaysRetry() error = %+v, wantErr %v", err, ret)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DoAlwaysRetry() got = %v, want %v", got, tt.want)
			}

			mockedObj.AssertNumberOfCalls(t, "MockDoAlwaysFunc", tt.wantDoFuncNumCalled)

			if tt.wantMaxDuration > 0 && duration > tt.wantMaxDuration {
				t.Errorf("DoAlwaysRetry() duration = %v, want less than %v", duration, tt.wantMaxDuration)
				return
			}

		})
	}
}

func TestNewRetriever(t *testing.T) {
	type args struct {
		config Config
	}
	tests := []struct {
		name string
		args args
		want Retriever
	}{
		{
			name: "when config is empty then return default config",
			args: args{
				config: Config{},
			},
			want: &retriever{
				maxAttempt:          3,
				maxTotalAttemptTime: 0,
				backoff:             NewExponentialBackoff(100, 2.0, 0.3),
				useGoroutine:        false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewRetriever(tt.args.config), "NewRetriever(%v)", tt.args.config)
		})
	}
}
