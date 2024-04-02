package httpcli

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/tenz-io/gokit/logger"
)

type trafficTransport struct {
	tripper http.RoundTripper
}

func (tt *trafficTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	var (
		ctx        = req.Context()
		url        = req.URL.Path
		reqData    []byte
		reqHeaders = req.Header
		le         = logger.FromContext(ctx)
	)

	reqData, err = httputil.DumpRequestOut(req, true)
	if err != nil {
		le.WithError(err).Warn("error dumping request")
		return tt.tripper.RoundTrip(req)
	}

	rec := logger.StartTrafficRec(ctx, &logger.ReqEntity{
		Typ: logger.TrafficTypSend,
		Cmd: url,
		Req: reqData,
		Fields: logger.Fields{
			"method":     req.Method,
			"query":      req.URL.Query(),
			"req_header": reqHeaders,
		},
	})

	defer func() {
		var (
			code        int
			msg         string
			respData    []byte
			respHeaders http.Header
			respErr     error
		)
		if err != nil {
			code = 1
			msg = err.Error()
		} else {
			code = resp.StatusCode
			respHeaders = resp.Header
			respData, respErr = httputil.DumpResponse(resp, true)
			if respErr != nil {
				msg = respErr.Error()
				le.WithError(respErr).Warn("error dumping response")
			}
		}

		rec.End(&logger.RespEntity{
			Code: fmt.Sprintf("%d", code),
			Msg:  msg,
			Resp: respData,
		}, logger.Fields{
			"resp_header": respHeaders,
		})
	}()

	return tt.tripper.RoundTrip(req)
}
