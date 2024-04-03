package httpcli

import "net/http"

type Interceptor interface {
	Apply(hc *http.Client)
}

type interceptor struct {
	config Config
}

func NewInterceptorWithOpts(opts ...ConfigOption) Interceptor {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	return NewInterceptor(config)
}

func NewInterceptor(config Config) Interceptor {
	return &interceptor{
		config: config,
	}
}

func (i *interceptor) Apply(hc *http.Client) {
	if hc == nil {
		return
	}

	transport := hc.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if i.config.EnableMetrics {
		transport = &metricsTransport{
			tripper: transport,
		}
	}

	if i.config.EnableTraffic {
		transport = &trafficTransport{
			tripper: transport,
		}
	}

	hc.Transport = transport
}
