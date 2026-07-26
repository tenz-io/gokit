package httpext

import (
	"fmt"
	"net/http"

	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/tracer/v3"
)

// trafficTransport records the request/response span via logger/v3's traffic
// logger. It records when EnableTraffic is set OR the request context carries
// tracer.FlagDebug (per-request debug traffic), so an operator can toggle
// capture for a single request without changing client config.
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
	ctx := req.Context()

	// Gate: capture only when configured on, or when the request context is in
	// debug mode (per-request opt-in). When off, this layer is a plain pass
	// through to the parent.
	if !tt.enable && !tracer.FromContext(ctx).IsDebug() {
		return tt.tripper.RoundTrip(req)
	}

	// Start the traffic span as a "send" (this is an outbound call). The span
	// records cmd/cost/code plus the structured fields below; it does NOT read
	// or record the request/response bodies — keep traffic logging cheap and
	// side-effect-free on the body streams.
	rec := logger.FromContext(ctx).StartTraffic(req.URL.Path).WithTyp(logger.TrafficTypSend)

	defer func() {
		// Defaults for the error/no-response path: code 1 is the sentinel for
		// "no status known"; resp_header stays nil.
		code := "1"
		var respHeaders http.Header
		switch {
		case err != nil:
			// Network/transport error before any response: end the span with
			// the error and the request-side fields.
			rec.EndWithError(err,
				"method", req.Method,
				"url", req.URL.String(),
				"query", req.URL.Query(),
				"req_header", req.Header,
			)
			return
		case resp != nil:
			code = fmt.Sprintf("%d", resp.StatusCode)
			respHeaders = resp.Header
		}

		rec.End(nil, code,
			"method", req.Method,
			"url", req.URL.String(),
			"query", req.URL.Query(),
			"req_header", req.Header,
			"resp_header", respHeaders,
		)
	}()

	return tt.tripper.RoundTrip(req)
}

// active reports whether this layer links into the chain. trafficTransport is
// always active once it has a parent: it must stay in the chain even when
// EnableTraffic is false, because a per-request context carrying
// tracer.FlagDebug must still be able to capture traffic. The enable flag
// only gates capture for non-debug requests (see the RoundTrip guard).
func (tt *trafficTransport) active() bool {
	if tt == nil || tt.tripper == nil {
		return false
	}
	return true
}
