package ginterceptor

import (
	"context"
	"fmt"
	syslog "log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/monitor"
)

var (
	newAppliers = []newApplierFunc{
		newTrackingApplier,
		newMetricsApplier,
		newTrafficApplier,
		newTimeoutApplier,
	}
)

type applier interface {
	active() bool
	apply() gin.HandlerFunc
}

type newApplierFunc func(config Config) applier

type metricsApplier struct {
	enable bool
}

func newMetricsApplier(config Config) applier {
	return &metricsApplier{
		enable: config.EnableMetrics,
	}
}

func (m *metricsApplier) active() bool {
	return m != nil && m.enable
}

func (m *metricsApplier) apply() gin.HandlerFunc {
	if !m.active() {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	syslog.Println("[gin-interceptor] apply metrics")

	return func(c *gin.Context) {
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

type timeoutApplier struct {
	timeout time.Duration
}

func newTimeoutApplier(config Config) applier {
	return &timeoutApplier{
		timeout: config.Timeout,
	}
}

func (t *timeoutApplier) active() bool {
	return t != nil && t.timeout > 0
}

func (t *timeoutApplier) apply() gin.HandlerFunc {
	if !t.active() {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	syslog.Println("[gin-interceptor] apply timeout:", t.timeout)

	return func(c *gin.Context) {
		var (
			ctx = c.Request.Context()
		)

		timeoutCtx, cancel := context.WithTimeout(ctx, t.timeout)
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
