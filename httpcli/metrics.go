package httpcli

import (
	"fmt"
	"net/http"

	"github.com/tenz-io/gokit/monitor"
)

type metricsTransport struct {
	tripper http.RoundTripper
}

func (mt *metricsTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	var (
		ctx = req.Context()
		url = req.URL.Path
	)

	rec := monitor.BeginRecord(ctx, url)

	defer func() {
		var (
			code int
		)

		if err != nil {
			code = 1
		} else {
			code = resp.StatusCode
		}

		rec.EndWithCode(fmt.Sprintf("%d", code))
	}()

	return mt.tripper.RoundTrip(req)
}
