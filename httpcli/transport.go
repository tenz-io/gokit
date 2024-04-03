package httpcli

import (
	"net/http"
)

var (
	transporters = []newTransporterFunc{
		newMetricsTransport,
		newTrafficTransport,
	}
)

type transporter interface {
	http.RoundTripper
	active() bool
}

type newTransporterFunc func(config Config, parent http.RoundTripper) transporter
