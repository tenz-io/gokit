package ginext

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/logger"
)

func setup(_ *testing.T) func() {
	logger.ConfigureWithOpts(
		logger.WithConsoleEnabled(true),
		logger.WithFileEnabled(false),
		logger.WithCallerEnabled(true),
		logger.WithLoggerLevel(logger.DebugLevel),
	)
	logger.ConfigureTrafficWithOpts(
		logger.WithTrafficEnabled(true),
	)

	return func() {
		time.Sleep(100 * time.Millisecond)
	}
}

func Test_captureRequest(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	type args struct {
		c *gin.Context
	}
	type after func(*testing.T, *args)
	tests := []struct {
		name    string
		args    args
		wantRes any
		after   after
	}{
		{
			name: "when context-type is application/json, return map[string]any",
			args: args{
				c: &gin.Context{
					Request: &http.Request{
						URL: &url.URL{
							Scheme: "https",
							Host:   "localhost",
							Path:   "/path",
						},
						Header: map[string][]string{
							"Content-Type": {"application/json"},
						},
						//ContentLength: 13,
						Body: func() io.ReadCloser {
							return io.NopCloser(
								strings.NewReader(`{"key":"value"}`),
							)
						}(),
					},
				},
			},
			wantRes: map[string]any{
				"key": "value",
			},
			after: func(t *testing.T, args *args) {
				// read request body
				bs, err := io.ReadAll(args.c.Request.Body)
				if err != nil {
					t.Errorf("error reading request body: %v", err)
					return
				}
				t.Logf("read content: %s", bs)

				if len(bs) == 0 {
					t.Errorf("request body is empty")
					return
				}

				if string(bs) != `{"key":"value"}` {
					t.Errorf("request body is not equal: %s != %s", string(bs), `{"key":"value"}`)
					return
				}

			},
		},
		{
			name: "when context-type is application/x-www-form-urlencoded, return map[string][]string",
			args: args{
				c: &gin.Context{
					Request: &http.Request{
						URL: &url.URL{
							Scheme: "https",
							Host:   "localhost",
							Path:   "/path",
						},
						Header: map[string][]string{
							"Content-Type": {"application/x-www-form-urlencoded"},
						},
						PostForm: map[string][]string{
							"key": {"value"},
						},
					},
				},
			},
			wantRes: url.Values{
				"key": {"value"},
			},
			after: func(t *testing.T, args *args) {
				form := args.c.Request.PostForm
				t.Logf("read content: %+v", form)

				if len(form) == 0 {
					t.Errorf("request postform is empty")
					return
				}

				if form.Get("key") != "value" {
					t.Errorf("request postform is not equal: %s != %s", form.Get("key"), "value")
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := captureRequest(tt.args.c); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("captureRequest() = %v, want %v", gotRes, tt.wantRes)
			}
			tt.after(t, &tt.args)
		})
	}
}
