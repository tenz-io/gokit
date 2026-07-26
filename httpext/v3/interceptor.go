package httpext

import "net/http"

// Interceptor composes a chain of http.RoundTripper layers (header injection,
// metrics, traffic) over a client's existing transport. Use NewInterceptor or
// NewInterceptorWithOpts to build one, then Apply it onto an *http.Client.
type Interceptor interface {
	// Intercept returns an http.RoundTripper that wraps tripper with every
	// active layer in newTransporters order. A nil tripper falls back to
	// http.DefaultTransport.
	Intercept(tripper http.RoundTripper) http.RoundTripper
	// Apply replaces hc.Transport with the intercepted chain. A nil hc is a
	// no-op.
	Apply(hc *http.Client)
}

type interceptor struct {
	config Config
}

// NewInterceptorWithOpts builds an interceptor from functional options layered
// over defaultConfig.
func NewInterceptorWithOpts(opts ...ConfigOption) Interceptor {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	return NewInterceptor(config)
}

// NewInterceptor builds an interceptor from an explicit Config.
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

	// Fold the chain inside-out: each active transporter wraps the running
	// transport, so the result is traffic(metrics(injectHeader(parent))).
	// Inactive layers are skipped so the chain contains only enabled layers.
	for _, newTransporter := range newTransporters {
		layer := newTransporter(i.config, transport)
		if layer.active() {
			transport = layer
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
