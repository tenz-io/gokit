package httpcli

import (
	"fmt"
	"net/http"

	"github.com/tenz-io/gokit/monitor"
)

type metricsTransport struct {
	enable  bool
	tripper http.RoundTripper
}

func newMetricsTransport(config Config, parent http.RoundTripper) transporter {
	return &metricsTransport{
		enable:  config.EnableMetrics,
		tripper: parent,
	}
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

func (mt *metricsTransport) active() bool {
	if mt == nil || mt.tripper == nil {
		return false
	}
	return mt.enable
}
