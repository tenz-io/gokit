package httpcli

import "net/http"

// Interceptor is an interface that wraps the Intercept and Apply methods.
type Interceptor interface {
	// Intercept returns a new http.RoundTripper that wraps the given tripper.
	Intercept(tripper http.RoundTripper) http.RoundTripper
	// Apply applies the interceptor to the given http.Client.
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

func (i *interceptor) Intercept(tripper http.RoundTripper) http.RoundTripper {
	transport := tripper
	if transport == nil {
		transport = http.DefaultTransport
	}

	for _, newTransporter := range newTransporters {
		tempTransport, ok := newTransporter(i.config, transport).(transporter)
		if ok && tempTransport.active() {
			transport = tempTransport
		}
	}

	return transport
}

func (i *interceptor) Apply(hc *http.Client) {
	if hc == nil {
		return
	}

	hc.Transport = i.Intercept(hc.Transport)
}
