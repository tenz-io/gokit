package httpcli

import "net/http"

type transport struct {
	tripper http.RoundTripper
}

// NewTransportWithOpts creates a new http.RoundTripper that wraps around an existing http.RoundTripper,
// allowing you to log requests and responses.
func NewTransportWithOpts(tripper http.RoundTripper, opts ...ConfigOption) http.RoundTripper {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}

	return NewTransport(tripper, config)
}

// NewTransport creates a new http.RoundTripper that wraps around an existing http.RoundTripper, allowing you to log requests and responses.
func NewTransport(tripper http.RoundTripper, config Config) http.RoundTripper {
	if tripper == nil {
		tripper = http.DefaultTransport
	}

	if config.EnableMetrics {
		tripper = &metricsTransport{
			tripper: tripper,
		}
	}

	if config.EnableTraffic {
		tripper = &trafficTransport{
			tripper: tripper,
		}
	}

	return &transport{
		tripper: tripper,
	}
}

func (t transport) RoundTrip(request *http.Request) (*http.Response, error) {
	return t.tripper.RoundTrip(request)
}
