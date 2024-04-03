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

	for _, newTransporter := range newTransporters {
		tempTransport, ok := newTransporter(i.config, transport).(transporter)
		if ok && tempTransport.active() {
			transport = tempTransport
		}
	}

	hc.Transport = transport
}
