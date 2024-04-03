package httpcli

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tenz-io/gokit/logger"
)

type mockMetricsTransport struct {
	mock.Mock
}

func (m *mockMetricsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func setup(_ *testing.T) (teardown func(*testing.T)) {
	logger.ConfigureWithOpts(
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithSetAsDefaultLvl(true),
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(false),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)

	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficConsoleEnabled(true),
		logger.WithTrafficFileEnabled(true),
	)

	return func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
		t.Logf("teardown")
	}
}

func Test_interceptor_Apply(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	type fields struct {
		config Config
	}
	type args struct {
		hc *http.Client
	}
	type after func(*testing.T, *fields, *args)
	tests := []struct {
		name   string
		fields fields
		args   args
		after  after
	}{
		{
			name: "when hc is nil then do nothing",
			fields: fields{
				config: Config{
					EnableMetrics: false,
					EnableTraffic: false,
				},
			},
			args: args{
				hc: nil,
			},
			after: func(t *testing.T, fields *fields, args *args) {
				if args.hc != nil {
					t.Errorf("interceptor.Apply() = %v, want %v", args.hc, nil)
					return
				}
			},
		},
		{
			name: "when hc transport is nil then set default transport",
			fields: fields{
				config: Config{
					EnableMetrics: false,
					EnableTraffic: false,
				},
			},
			args: args{
				hc: &http.Client{},
			},
			after: func(t *testing.T, fields *fields, args *args) {
				if args.hc.Transport == nil {
					t.Errorf("interceptor.Apply() = %v, nil is not expected", args.hc.Transport)
					return
				}

				if !reflect.DeepEqual(args.hc.Transport, http.DefaultTransport) {
					t.Errorf("interceptor.Apply() = %v, want %v", args.hc.Transport, http.DefaultTransport)
					return
				}

			},
		},
		{
			name: "when hc transport is not nil then do nothing",
			fields: fields{
				config: Config{
					EnableMetrics: false,
					EnableTraffic: false,
				},
			},
			args: args{
				hc: &http.Client{
					Transport: &mockMetricsTransport{},
				},
			},
			after: func(t *testing.T, fields *fields, args *args) {
				if args.hc.Transport == nil {
					t.Errorf("interceptor.Apply() = %v, nil is not expected", args.hc.Transport)
					return
				}

				if !reflect.DeepEqual(args.hc.Transport, &mockMetricsTransport{}) {
					t.Errorf("interceptor.Apply() = %v, want %v", args.hc.Transport, &mockMetricsTransport{})
					return
				}

			},
		},
		{
			name: "when hc transport is not nil and apply metrics transport",
			fields: fields{
				config: Config{
					EnableMetrics: true,
					EnableTraffic: false,
				},
			},
			args: args{
				hc: &http.Client{
					Transport: &mockMetricsTransport{},
				},
			},
			after: func(t *testing.T, fields *fields, args *args) {
				if args.hc.Transport == nil {
					t.Errorf("interceptor.Apply() = %v, nil is not expected", args.hc.Transport)
					return
				}

				transport, ok := args.hc.Transport.(*metricsTransport)
				if !ok {
					t.Errorf("interceptor.Apply() = %v, want %v", args.hc.Transport, &metricsTransport{})
					return
				}

				if !reflect.DeepEqual(transport.tripper, &mockMetricsTransport{}) {
					t.Errorf("interceptor.Apply() = %v, want %v", transport.tripper, &mockMetricsTransport{})
					return
				}
			},
		},
		{
			name: "when hc transport is not nil and apply traffic transport",
			fields: fields{
				config: Config{
					EnableMetrics: false,
					EnableTraffic: true,
				},
			},
			args: args{
				hc: &http.Client{
					Transport: &mockMetricsTransport{},
				},
			},
			after: func(t *testing.T, fields *fields, args *args) {
				if args.hc.Transport == nil {
					t.Errorf("interceptor.Apply() = %v, nil is not expected", args.hc.Transport)
					return
				}

				transport, ok := args.hc.Transport.(*trafficTransport)
				if !ok {
					t.Errorf("interceptor.Apply() = %v, want %v", args.hc.Transport, &trafficTransport{})
					return
				}

				if !reflect.DeepEqual(transport.tripper, &mockMetricsTransport{}) {
					t.Errorf("interceptor.Apply() = %v, want %v", transport.tripper, &mockMetricsTransport{})
					return
				}
			},
		},
		{
			name: "when hc transport is not nil and apply metrics and traffic transport",
			fields: fields{
				config: Config{
					EnableMetrics: true,
					EnableTraffic: true,
				},
			},
			args: args{
				hc: &http.Client{
					Transport: &mockMetricsTransport{},
				},
			},
			after: func(t *testing.T, fields *fields, args *args) {
				if args.hc.Transport == nil {
					t.Errorf("interceptor.Apply() = %v, nil is not expected", args.hc.Transport)
					return
				}

				transport, ok := args.hc.Transport.(*trafficTransport)
				if !ok {
					t.Errorf("interceptor.Apply() = %v, want %v", args.hc.Transport, &trafficTransport{})
					return
				}

				mt, ok := transport.tripper.(*metricsTransport)
				if !ok {
					t.Errorf("interceptor.Apply() = %v, want %v", args.hc.Transport, &metricsTransport{})
					return
				}

				if !reflect.DeepEqual(mt.tripper, &mockMetricsTransport{}) {
					t.Errorf("interceptor.Apply() = %v, want %v", transport.tripper, &mockMetricsTransport{})
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &interceptor{
				config: tt.fields.config,
			}
			i.Apply(tt.args.hc)
			tt.after(t, &tt.fields, &tt.args)
		})
	}
}

func TestInterceptor(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	type fields struct {
		config Config
		hc     *http.Client
	}
	type args struct {
		req *http.Request
	}
	type before func(*testing.T, *fields, *args)
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp bool
		wantErr  assert.ErrorAssertionFunc
		before   before
	}{
		{
			name: "when metrics and traffic are disabled",
			fields: fields{
				config: Config{
					EnableMetrics: false,
					EnableTraffic: false,
				},
				hc: &http.Client{
					Transport: &mockMetricsTransport{},
				},
			},
			args: args{
				req: &http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost",
						Path:   "/",
					},
					Header: make(http.Header),
				},
			},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, fields *fields, args *args) {
				var (
					mockedTransport = fields.hc.Transport.(*mockMetricsTransport)
				)

				mockedTransport.On("RoundTrip", args.req).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						return io.NopCloser(strings.NewReader("hello world"))
					}(),
				}, nil).Times(1)
			},
		},
		{
			name: "when metrics is enabled",
			fields: fields{
				config: Config{
					EnableMetrics: true,
					EnableTraffic: false,
				},
				hc: &http.Client{
					Transport: &mockMetricsTransport{},
				},
			},
			args: args{
				req: &http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost",
						Path:   "/",
					},
					Header: make(http.Header),
				},
			},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, fields *fields, args *args) {

				// iterate get deep level of transport until we get the mocked transport
				var (
					mockedTransport = fields.hc.Transport.(*metricsTransport).tripper.(*mockMetricsTransport)
				)

				mockedTransport.On("RoundTrip", args.req).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body: func() io.ReadCloser {
						return io.NopCloser(strings.NewReader("hello world"))
					}(),
				}, nil).Times(1)
			},
		},
		{
			name: "when traffic is enabled",
			fields: fields{
				config: Config{
					EnableMetrics: false,
					EnableTraffic: true,
				},
				hc: &http.Client{
					Transport: &mockMetricsTransport{},
				},
			},
			args: args{
				req: &http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost",
						Path:   "/",
					},
					Header: make(http.Header),
				},
			},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, fields *fields, args *args) {
				// iterate get deep level of transport until we get the mocked transport
				var (
					mockedTransport = fields.hc.Transport.(*trafficTransport).tripper.(*mockMetricsTransport)
				)

				mockedTransport.On("RoundTrip", args.req).Return(&http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/plain; charset=utf-8"},
					},
					Body: func() io.ReadCloser {
						return io.NopCloser(strings.NewReader("hello world"))
					}(),
				}, nil).Times(1)
			},
		},
		{
			name: "when traffic and metrics is enabled",
			fields: fields{
				config: Config{
					EnableMetrics: true,
					EnableTraffic: true,
				},
				hc: &http.Client{
					Transport: &mockMetricsTransport{},
				},
			},
			args: args{
				req: &http.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost",
						Path:   "/hello",
					},
					Header: http.Header{
						"Content-Type": []string{"text/plain; charset=utf-8"},
					},
					Body: func() io.ReadCloser {
						return io.NopCloser(strings.NewReader("hello world"))
					}(),
				},
			},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, fields *fields, args *args) {
				// iterate get deep level of transport until we get the mocked transport
				var (
					mockedTransport = fields.hc.Transport.(*trafficTransport).tripper.(*metricsTransport).tripper.(*mockMetricsTransport)
				)

				mockedTransport.On("RoundTrip", args.req).Return(&http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{"text/plain; charset=utf-8"},
					},
					Body: func() io.ReadCloser {
						return io.NopCloser(strings.NewReader("hello world"))
					}(),
				}, nil).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewInterceptor(tt.fields.config)
			i.Apply(tt.fields.hc)
			tt.before(t, &tt.fields, &tt.args)

			resp, err := tt.fields.hc.Do(tt.args.req)
			if !tt.wantErr(t, err) {
				t.Errorf("Interceptor() error = %v, is not expected", err)
			}

			if tt.wantResp && resp == nil {
				t.Errorf("Interceptor() resp = %v, want not nil", resp)
			}

			if resp != nil {
				respCopy, err := httputil.DumpResponse(resp, true)
				if err != nil {
					t.Logf("dump response err: %v", err)
					return
				}
				t.Logf("resp: %s", respCopy)
			}

		})
	}

}
