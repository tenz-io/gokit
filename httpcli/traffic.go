package httpcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tenz-io/gokit/logger"
)

type trafficTransport struct {
	enable  bool
	tripper http.RoundTripper
}

func newTrafficTransport(config Config, parent http.RoundTripper) transporter {
	return &trafficTransport{
		enable:  config.EnableTraffic,
		tripper: parent,
	}
}

func (tt *trafficTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	var (
		ctx        = req.Context()
		url        = req.URL.Path
		reqHeaders = req.Header
	)

	rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
		Typ: logger.TrafficTypSend,
		Cmd: url,
		Req: captureRequest(req),
		Fields: logger.Fields{
			"method":     req.Method,
			"query":      req.URL.Query(),
			"req_header": reqHeaders,
		},
	})

	defer func() {
		var (
			code        = 1
			respHeaders http.Header
		)
		if err == nil && resp != nil {
			code = resp.StatusCode
			respHeaders = resp.Header
		}

		rec.End(&logger.RespEntity{
			Code: fmt.Sprintf("%d", code),
			Msg:  errorMsg(err),
			Resp: captureResponse(resp),
		}, logger.Fields{
			"resp_header": respHeaders,
		})
	}()

	return tt.tripper.RoundTrip(req)
}

func (tt *trafficTransport) active() bool {
	if tt == nil || tt.tripper == nil {
		return false
	}
	return tt.enable
}

func errorMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func captureRequest(req *http.Request) (res any) {
	if req == nil {
		return nil
	}

	var (
		body        []byte
		err         error
		contentType = strings.ToLower(req.Header.Get("Content-Type"))
		ctx         = req.Context()
	)

	le := logger.FromContext(ctx).WithFields(logger.Fields{
		"Content-Type": contentType,
	})

	if req.Method == http.MethodGet {
		le.Debug("GET method request, skip capture request")
		return nil
	}

	if req.Body == nil {
		le.Debug("request body is nil")
		return nil
	}

	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		return req.PostForm
	}

	if strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "text/") {
		body, err = io.ReadAll(req.Body)
		if err != nil {
			le.WithError(err).Warn("error reading request body")
			return nil
		}

		// clone body for reset body
		bs := bytes.Clone(body)
		defer func() {
			req.Body = io.NopCloser(bytes.NewBuffer(bs))
		}()
	} else {
		le.Debug("unsupported dump content-type")
		return "<unsupported content-type>"
	}

	if len(body) == 0 {
		le.Debug("request body is empty")
		return nil
	}

	if strings.HasPrefix(contentType, "application/json") {
		var reqMap map[string]any
		if err = json.Unmarshal(body, &reqMap); err != nil {
			le.WithError(err).Warnf("json unmarshal request failed")
			return "<json unmarshal failed>"
		}

		return reqMap
	}

	// return string for other content-type
	return limitString(string(body), 128)
}

// captureResponse capture response from http response
func captureResponse(resp *http.Response) any {
	if resp == nil {
		return nil
	}

	var (
		ctx = ifThen(resp.Request != nil, func() context.Context {
			return resp.Request.Context()
		}, func() context.Context {
			return context.Background()
		}).(context.Context)
		contentType = strings.ToLower(resp.Header.Get("Content-Type"))
		body        []byte
		err         error
		le          = logger.FromContext(ctx).WithFields(logger.Fields{
			"Content-Type": contentType,
		})
	)

	if resp.Body == nil {
		le.Debug("response body is nil")
		return nil
	}

	if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/") {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			le.WithError(err).Warn("error reading response body")
			return nil
		}

		// clone body for reset body
		bodyCopy := bytes.Clone(body)
		defer func() {
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyCopy))
		}()
	} else {
		le.Debug("unsupported dump content-type")
		return "<unsupported content-type>"
	}

	if len(body) == 0 {
		le.Debug("response body is empty")
		return nil
	}

	if strings.HasPrefix(contentType, "application/json") {
		var res map[string]any
		if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
			le.WithError(err).Warn("error decoding response body")
			return "<json decode failed>"
		}

		return res
	}

	le.Debug("capture response")
	return limitString(string(body), 128)
}

func ifThen[T any](cond bool, tf, ff func() T) T {
	if cond {
		return tf()
	}
	return ff()
}

// limitString returns a string with a maximum length of n.
func limitString(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
