package httpext

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tenz-io/gokit/logger/v3"
)

type mockTransport struct {
	mock.Mock
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *mockTransport) active() bool { return true }

func setup(t *testing.T) func(*testing.T) {
	t.Helper()
	logger.ConfigureWithOpts(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithFilePath(""),
		logger.WithCaller(true),
		logger.WithCallerSkip(1),
		logger.WithTraffic(true),
	)
	return func(t *testing.T) {
		// Traffic writes asynchronously via a lumberjack sync; give it a beat
		// to flush before the process exits.
		time.Sleep(100 * time.Millisecond)
		t.Logf("teardown")
	}
}

// Test_interceptor_Apply verifies the chain is wired (or not) per config.
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
			name:   "when hc is nil then do nothing",
			fields: fields{config: Config{}},
			args:   args{hc: nil},
			after: func(t *testing.T, f *fields, a *args) {
				assert.Nil(t, a.hc)
			},
		},
		{
			name:   "when hc transport is nil then set default transport",
			fields: fields{config: Config{}},
			args:   args{hc: &http.Client{}},
			after: func(t *testing.T, f *fields, a *args) {
				assert.NotNil(t, a.hc.Transport)
			},
		},
		{
			name:   "when nothing enabled then traffic still wraps parent (per-request debug)",
			fields: fields{config: Config{EnableMetrics: false, EnableTraffic: false}},
			args:   args{hc: &http.Client{Transport: &mockTransport{}}},
			after: func(t *testing.T, f *fields, a *args) {
				// trafficTransport is always active so a per-request FlagDebug
				// context can still capture; its internal enable flag only
				// gates non-debug traffic. So the chain is traffic(parent).
				tt, ok := a.hc.Transport.(*trafficTransport)
				assert.True(t, ok, "expected trafficTransport, got %T", a.hc.Transport)
				_, ok = tt.tripper.(*mockTransport)
				assert.True(t, ok, "expected parent mockTransport, got %T", tt.tripper)
				assert.False(t, tt.enable, "enable flag should be false")
			},
		},
		{
			name:   "when metrics enabled then traffic wraps metrics wraps parent",
			fields: fields{config: Config{EnableMetrics: true, EnableTraffic: false}},
			args:   args{hc: &http.Client{Transport: &mockTransport{}}},
			after: func(t *testing.T, f *fields, a *args) {
				tt, ok := a.hc.Transport.(*trafficTransport)
				assert.True(t, ok, "expected trafficTransport, got %T", a.hc.Transport)
				mt, ok := tt.tripper.(*metricsTransport)
				assert.True(t, ok, "expected metricsTransport, got %T", tt.tripper)
				_, ok = mt.tripper.(*mockTransport)
				assert.True(t, ok, "expected parent mockTransport, got %T", mt.tripper)
			},
		},
		{
			name:   "when traffic enabled then traffic wraps parent",
			fields: fields{config: Config{EnableMetrics: false, EnableTraffic: true}},
			args:   args{hc: &http.Client{Transport: &mockTransport{}}},
			after: func(t *testing.T, f *fields, a *args) {
				tt, ok := a.hc.Transport.(*trafficTransport)
				assert.True(t, ok, "expected trafficTransport, got %T", a.hc.Transport)
				_, ok = tt.tripper.(*mockTransport)
				assert.True(t, ok, "expected parent mockTransport, got %T", tt.tripper)
			},
		},
		{
			name:   "when metrics and traffic enabled then traffic wraps metrics wraps parent",
			fields: fields{config: Config{EnableMetrics: true, EnableTraffic: true}},
			args:   args{hc: &http.Client{Transport: &mockTransport{}}},
			after: func(t *testing.T, f *fields, a *args) {
				tt, ok := a.hc.Transport.(*trafficTransport)
				assert.True(t, ok, "expected trafficTransport, got %T", a.hc.Transport)
				mt, ok := tt.tripper.(*metricsTransport)
				assert.True(t, ok, "expected metricsTransport, got %T", tt.tripper)
				_, ok = mt.tripper.(*mockTransport)
				assert.True(t, ok, "expected parent mockTransport, got %T", mt.tripper)
			},
		},
		{
			name: "when headers configured then injectHeader is innermost",
			fields: fields{config: Config{
				EnableMetrics: true,
				EnableTraffic: true,
				Headers:       map[string]string{"Authorization": "Bearer token"},
			}},
			args: args{hc: &http.Client{Transport: &mockTransport{}}},
			after: func(t *testing.T, f *fields, a *args) {
				tt, ok := a.hc.Transport.(*trafficTransport)
				assert.True(t, ok, "expected trafficTransport, got %T", a.hc.Transport)
				mt, ok := tt.tripper.(*metricsTransport)
				assert.True(t, ok, "expected metricsTransport, got %T", tt.tripper)
				it, ok := mt.tripper.(*injectHeaderTransport)
				assert.True(t, ok, "expected injectHeaderTransport, got %T", mt.tripper)
				_, ok = it.tripper.(*mockTransport)
				assert.True(t, ok, "expected parent mockTransport, got %T", it.tripper)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &interceptor{config: tt.fields.config}
			i.Apply(tt.args.hc)
			tt.after(t, &tt.fields, &tt.args)
		})
	}
}

// TestInterceptor drives a real request through the chain end-to-end.
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
			name: "when nothing enabled then pass through",
			fields: fields{
				config: Config{EnableMetrics: false, EnableTraffic: false},
				hc:     &http.Client{Transport: &mockTransport{}},
			},
			args: args{req: &http.Request{
				Method: "GET",
				URL:    &url.URL{Scheme: "https", Host: "localhost", Path: "/"},
				Header: make(http.Header),
			}},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, f *fields, a *args) {
				// traffic is always outermost; peel it to reach the parent mock.
				mt := f.hc.Transport.(*trafficTransport).tripper.(*mockTransport)
				mt.On("RoundTrip", a.req).Return(&http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("hello world")),
				}, nil).Times(1)
			},
		},
		{
			name: "when metrics enabled",
			fields: fields{
				config: Config{EnableMetrics: true, EnableTraffic: false},
				hc:     &http.Client{Transport: &mockTransport{}},
			},
			args: args{req: &http.Request{
				Method: "GET",
				URL:    &url.URL{Scheme: "https", Host: "localhost", Path: "/"},
				Header: make(http.Header),
			}},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, f *fields, a *args) {
				// traffic(metrics(mock))
				tt0 := f.hc.Transport.(*trafficTransport)
				mt0 := tt0.tripper.(*metricsTransport)
				mt := mt0.tripper.(*mockTransport)
				mt.On("RoundTrip", a.req).Return(&http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("hello world")),
				}, nil).Times(1)
			},
		},
		{
			name: "when traffic enabled",
			fields: fields{
				config: Config{EnableMetrics: false, EnableTraffic: true},
				hc:     &http.Client{Transport: &mockTransport{}},
			},
			args: args{req: &http.Request{
				Method: "GET",
				URL:    &url.URL{Scheme: "https", Host: "localhost", Path: "/"},
				Header: make(http.Header),
			}},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, f *fields, a *args) {
				mt := f.hc.Transport.(*trafficTransport).tripper.(*mockTransport)
				mt.On("RoundTrip", a.req).Return(&http.Response{
					StatusCode: 200,
					Header:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
					Body:       io.NopCloser(strings.NewReader("hello world")),
				}, nil).Times(1)
			},
		},
		{
			name: "when traffic and metrics enabled with json body",
			fields: fields{
				config: Config{EnableMetrics: true, EnableTraffic: true},
				hc:     &http.Client{Transport: &mockTransport{}},
			},
			args: args{req: &http.Request{
				Method: "POST",
				URL:    &url.URL{Scheme: "https", Host: "localhost", Path: "/hello"},
				Header: http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
				Body:   io.NopCloser(strings.NewReader("hello world")),
			}},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, f *fields, a *args) {
				tt0 := f.hc.Transport.(*trafficTransport)
				mt0 := tt0.tripper.(*metricsTransport)
				mt := mt0.tripper.(*mockTransport)
				mt.On("RoundTrip", a.req).Return(&http.Response{
					StatusCode: 200,
					Header:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
					Body:       io.NopCloser(strings.NewReader("hello world")),
				}, nil).Times(1)
			},
		},
		{
			name: "when headers injected",
			fields: fields{
				config: Config{
					EnableMetrics: false,
					EnableTraffic: true,
					Headers: map[string]string{
						"Authorization": "Bearer token",
						"Content-Type":  "application/json",
						"X-Request-ID":  "123",
					},
				},
				hc: &http.Client{Transport: &mockTransport{}},
			},
			args: args{req: &http.Request{
				Method: "POST",
				URL:    &url.URL{Scheme: "https", Host: "localhost", Path: "/hello"},
				Header: http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
				Body:   io.NopCloser(strings.NewReader("hello world")),
			}},
			wantResp: true,
			wantErr:  assert.NoError,
			before: func(t *testing.T, f *fields, a *args) {
				tt0 := f.hc.Transport.(*trafficTransport)
				ih := tt0.tripper.(*injectHeaderTransport)
				mt := ih.tripper.(*mockTransport)
				mt.On("RoundTrip", a.req).Return(&http.Response{
					StatusCode: 200,
					Header:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}},
					Body:       io.NopCloser(strings.NewReader("hello world")),
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
			tt.wantErr(t, err)

			if tt.wantResp {
				assert.NotNil(t, resp)
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
