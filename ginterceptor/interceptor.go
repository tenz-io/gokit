package ginterceptor

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tenz-io/gokit/monitor"
	syslog "log"
	"net/http"
)

type Interceptor interface {
	ApplyTracking() gin.HandlerFunc
	ApplyTraffic() gin.HandlerFunc
	ApplyMetrics() gin.HandlerFunc
	ApplyTimeout() gin.HandlerFunc
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

type interceptor struct {
	config Config
}

func (i *interceptor) ApplyMetrics() gin.HandlerFunc {
	if !i.config.EnableMetrics {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	syslog.Println("[gin-interceptor] apply metrics")

	return func(c *gin.Context) {
		// get context from gin
		var (
			ctx = c.Request.Context()
		)
		rec := monitor.BeginRecord(ctx, "total")
		defer func() {
			httpStatus := c.Writer.Status()
			rec.EndWithCode(fmt.Sprintf("%d", httpStatus))
		}()

		c.Next()
	}
}

func (i *interceptor) ApplyTimeout() gin.HandlerFunc {
	if i.config.Timeout <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	syslog.Println("[gin-interceptor] apply timeout:", i.config.Timeout)

	return func(c *gin.Context) {
		var (
			ctx = c.Request.Context()
		)

		timeoutCtx, cancel := context.WithTimeout(ctx, i.config.Timeout)
		defer cancel()

		doneC := make(chan struct{})
		go func() {
			c.Next()
			close(doneC)
		}()

		select {
		case <-timeoutCtx.Done():
			c.AbortWithStatus(http.StatusRequestTimeout)
			return
		case <-doneC:
			// The request completed before the timeout
		}
	}
}
