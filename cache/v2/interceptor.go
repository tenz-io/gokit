package cache

import "github.com/go-redis/redis/v8"

type Interceptor interface {
	Apply(client *redis.Client)
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

func (i *interceptor) Apply(client *redis.Client) {
	for _, hook := range newHooks {
		client.AddHook(hook(i.config))
	}
}
