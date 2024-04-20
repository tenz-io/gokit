package ginext

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
	for _, newApplier := range newAppliers {
		a := newApplier(config)
		if a.active() {
			applierList = append(applierList, a)
		}
	}
	return &interceptor{
		appliers: applierList,
	}
}

type interceptor struct {
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
