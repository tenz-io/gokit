package httpext

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tenz-io/gokit/logger"
)

func prepare(t *testing.T) (teardown func(t *testing.T)) {
	logger.ConfigureWithOpts(
		logger.WithLoggerLevel(logger.DebugLevel),
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(true),
		logger.WithCallerEnabled(true),
		logger.WithCallerSkip(1),
	)

	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
		logger.WithTrafficIgnoresOpt(HeaderNameAuthorization),
	)

	return func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
	}
}

func Test_client_JSON(t *testing.T) {
	clean := prepare(t)
	defer clean(t)

	type fields struct {
		cli *http.Client
	}
	type args struct {
		ctx      context.Context
		url      string
		method   HttpMethod
		reqBody  any
		respBody any
		reqOpts  []RequestOption
	}
	type behavior func(*testing.T, *fields, *args)
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp any
		wantErr  assert.ErrorAssertionFunc
		behavior behavior
	}{
		{
			name: "when request is invalid then return error",
			fields: fields{
				cli: new(http.Client),
			},
			args: args{
				ctx:      context.Background(),
				url:      "",
				method:   MethodPost,
				reqBody:  nil,
				respBody: nil,
				reqOpts:  []RequestOption{},
			},
			wantErr:  assert.Error,
			wantResp: nil,
			behavior: func(t *testing.T, fields *fields, args *args) {
				// do nothing
			},
		},
		{
			name: "when request is valid then return nil",
			fields: fields{
				cli: func() *http.Client {
					cli := &http.Client{
						Transport: &mockTransport{},
					}

					icpt := NewInterceptorWithOpts(
						WithEnableMetrics(false),
						WithEnableTraffic(true),
						WithHeaders(map[string]string{"Authorization": "Bearer token"}),
					)
					icpt.Apply(cli)

					return cli
				}(),
			},
			args: args{
				ctx:      context.Background(),
				url:      "https://example.com/add",
				method:   MethodPost,
				reqBody:  map[string]int{"a": 1, "b": 2},
				respBody: &map[string]int{},
				reqOpts:  []RequestOption{},
			},
			wantResp: &map[string]int{"sum": 3},
			wantErr:  assert.NoError,
			behavior: func(t *testing.T, fields *fields, args *args) {
				var (
					mockedTransport = fields.cli.Transport.(*trafficTransport).
						tripper.(*injectHeaderTransport).
						tripper.(*mockTransport)
				)

				mockedTransport.On("RoundTrip", mock.Anything).Return(
					&http.Response{
						StatusCode: http.StatusOK,
						Header: http.Header{
							"Content-Type": []string{"application/json"},
						},
						Body: func() io.ReadCloser {
							return io.NopCloser(strings.NewReader(`{"sum": 3}`))
						}(),
					}, nil).Times(1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{
				cli: tt.fields.cli,
			}
			tt.behavior(t, &tt.fields, &tt.args)
			err := c.JSON(tt.args.ctx, tt.args.url, tt.args.method, tt.args.reqBody, tt.args.respBody, tt.args.reqOpts...)
			t.Logf("err: %v, req: %+v, resp: %+v", err, tt.args.reqBody, tt.args.respBody)
			tt.wantErr(t, err, fmt.Sprintf("JSON(%v, %v, %v, %v, %v, %v)",
				tt.args.ctx, tt.args.url, tt.args.method, tt.args.reqBody, tt.args.respBody, tt.args.reqOpts))

			if !assert.Equal(t, tt.args.respBody, tt.wantResp) {
				t.Errorf("want: %v, got: %v", tt.wantResp, tt.args.respBody)
			}

		})
	}
}
