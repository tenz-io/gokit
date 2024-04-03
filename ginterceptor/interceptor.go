package ginterceptor

import (
	"github.com/gin-gonic/gin"
)

type Interceptor interface {
	Apply(engine *gin.Engine)
}

func NewInterceptorWithOpts(opts ...ConfigOption) Interceptor {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}
	return NewInterceptor(config)

}

func NewInterceptor(config Config) Interceptor {
	var applierList []applier
	for _, newApplier := range appliers {
		a := newApplier(config)
		if a.active() {
			applierList = append(applierList, a)
		}
	}
	return &interceptor{
		config:   config,
		appliers: applierList,
	}
}

type interceptor struct {
	config   Config
	appliers []applier
}

func (i *interceptor) Apply(engine *gin.Engine) {
	if engine == nil {
		return
	}

	for _, a := range i.appliers {
		engine.Use(a.apply())
	}

}
