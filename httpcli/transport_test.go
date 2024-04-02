package httpcli

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tenz-io/gokit/logger"
)

type mockTransport struct {
	mock.Mock
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		Mock: mock.Mock{},
	}
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func setup(_ *testing.T) (tearDown func()) {
	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficConsoleEnabled(true),
		logger.WithTrafficFileEnabled(false),
	)
	return func() {
		time.Sleep(100 * time.Millisecond)
	}
}

func TestRoundTrip(t *testing.T) {

	tearDown := setup(t)
	defer tearDown()

	type fields struct {
		tripper http.RoundTripper
		opts    []ConfigOption
	}
	type args struct {
		request *http.Request
	}
	type behavior func(*fields, *args)
	tests := []struct {
		name     string
		fields   fields
		args     args
		behavior behavior
		wantResp *http.Response
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "when enabled metrics and traffic then return response",
			fields: fields{
				tripper: newMockTransport(),
				opts: []ConfigOption{
					WithEnableMetrics(true),
					WithEnableTraffic(true),
				},
			},
			args: args{
				request: &http.Request{
					URL: &url.URL{
						Path:   "/index",
						Scheme: "https",
						Host:   "localhost",
					},
					Header: make(http.Header),
				},
			},
			behavior: func(fields *fields, args *args) {
				var (
					mockedTransport = fields.tripper.(*mockTransport)
				)
				mockedTransport.On("RoundTrip", args.request).
					Return(&http.Response{
						StatusCode: 200,
					}, nil).
					Times(1)
			},
			wantResp: &http.Response{
				StatusCode: 200,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tripper := NewTransportWithOpts(tt.fields.tripper, tt.fields.opts...)
			tt.behavior(&tt.fields, &tt.args)
			gotResp, gotErr := tripper.RoundTrip(tt.args.request)

			if !tt.wantErr(t, gotErr) {
				t.Errorf("RoundTrip() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantResp, gotResp)

		})
	}

}
