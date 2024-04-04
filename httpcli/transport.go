package httpcli

import (
	"fmt"
	"net/http"

	"github.com/tenz-io/gokit/monitor"
)

var (
	newTransporters = []newTransporterFunc{
		newInjectHeaderTransport,
		newMetricsTransport,
		newTrafficTransport,
	}
)

const (
	HeaderNameAuthorization = "Authorization"
	HeaderNameContentType   = "Content-Type"
)

type transporter interface {
	http.RoundTripper
	active() bool
}

type newTransporterFunc func(config Config, parent http.RoundTripper) transporter

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
		ctx  = req.Context()
		url  = req.URL.Path
		code = 1
	)

	rec := monitor.BeginRecord(ctx, url)

	defer func() {
		if err == nil {
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

type injectHeaderTransport struct {
	tripper http.RoundTripper
	headers map[string]string
}

func newInjectHeaderTransport(config Config, parent http.RoundTripper) transporter {
	return &injectHeaderTransport{
		tripper: parent,
		headers: config.Headers,
	}
}

func (it *injectHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range it.headers {
		req.Header.Set(key, value)
	}

	return it.tripper.RoundTrip(req)
}

func (it *injectHeaderTransport) active() bool {
	return it != nil && it.tripper != nil && len(it.headers) > 0
}
