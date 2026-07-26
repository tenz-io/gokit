package httpext

import (
	"fmt"
	"net/http"

	"github.com/tenz-io/gokit/monitor/v3"
)

// newTransporters is the ordered chain builder. Order is outermost-first:
// injectHeader runs before metrics runs before traffic, so the final
// http.RoundTripper is traffic(metrics(injectHeader(parent))). Each layer only
// links in when its active() reports true, so a disabled layer is a no-op pass
// through rather than a dead wrapper.
var newTransporters = []newTransporterFunc{
	newInjectHeaderTransport,
	newMetricsTransport,
	newTrafficTransport,
}

// HeaderName constants name the headers this package sets or reads.
const (
	HeaderNameAuthorization = "Authorization"
	HeaderNameContentType   = "Content-Type"
)

// transporter is the internal contract a chain layer satisfies: it wraps a
// parent RoundTripper and reports whether it is active. An inactive layer is
// dropped from the chain so callers (and tests) see a flat chain of only the
// enabled layers.
type transporter interface {
	http.RoundTripper
	active() bool
}

// newTransporterFunc builds a chain layer from config over parent.
type newTransporterFunc func(config Config, parent http.RoundTripper) transporter

// metricsTransport records per-request latency and a status code counter via
// monitor/v3. It is a no-op when EnableMetrics is false or when the context
// carries no Exporter (monitor.Begin returns a nil-safe Recorder then).
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
		code = "1" // sentinel: 1 means the call errored before a status was known
	)

	rec := monitor.Begin(ctx, url)

	defer func() {
		if err == nil && resp != nil {
			code = fmt.Sprintf("%d", resp.StatusCode)
		}
		rec.EndWithCode(code)
	}()

	return mt.tripper.RoundTrip(req)
}

func (mt *metricsTransport) active() bool {
	if mt == nil || mt.tripper == nil {
		return false
	}
	return mt.enable
}

// injectHeaderTransport sets the configured headers on every outbound request
// before delegating to the parent. Header.Set replaces, so a caller that sets
// the same header on a specific request still loses to the configured value —
// configure only headers that should be uniform across all calls.
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
