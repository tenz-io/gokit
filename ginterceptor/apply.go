package ginterceptor

import (
	"context"
	"fmt"
	syslog "log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
)

var (
	newAppliers = []newApplierFunc{
		newAccessLogApplier,
		newSlowLogApplier,
		newTrackingApplier,
		newMetricsApplier,
		newTrafficApplier,
		newTimeoutApplier,
		newPanicRecoveryApplier,
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
		}
	}
}

type accessLogApplier struct {
	enable    bool
	accessLog string
}

func newAccessLogApplier(config Config) applier {
	return &accessLogApplier{
		enable:    config.EnableAccessLog,
		accessLog: config.AccessLog,
	}
}

func (a *accessLogApplier) active() bool {
	return a != nil && a.enable
}

func (a *accessLogApplier) apply() gin.HandlerFunc {
	if !a.active() {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	accessLog := a.accessLog
	if accessLog == "" {
		accessLog = "log"
	}

	filename := strings.Join([]string{accessLog, "access.log"}, "/")

	syslog.Println("[gin-interceptor] apply access log:", filename)

	accessLogger := &lumberjack.Logger{
		Filename:   filename,
		LocalTime:  true,
		MaxSize:    10,   // the maximum size of each log file (in megabytes)
		MaxBackups: 5,    // the maximum number of old log files to retain
		MaxAge:     30,   // the maximum number of days to retain old log files
		Compress:   true, // compress old log files with gzip
	}

	return gin.LoggerWithWriter(accessLogger)
}

type panicRecoveryApplier struct {
}

func newPanicRecoveryApplier(_ Config) applier {
	return &panicRecoveryApplier{}
}

func (p *panicRecoveryApplier) active() bool {
	return true
}

func (p *panicRecoveryApplier) apply() gin.HandlerFunc {
	syslog.Println("[gin-interceptor] apply panic recovery")
	return func(c *gin.Context) {
		var (
			ctx = c.Request.Context()
		)
		defer func() {
			if r := recover(); r != nil {
				syslog.Printf("panic recovery: %s, stacktrace: %s\n", r, string(debug.Stack()))
				logger.FromContext(ctx).WithFields(logger.Fields{
					"panic": fmt.Sprintf("%s", r),
				}).Error("panic recovery")
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}

type slowLogApplier struct {
	slowLogFloor time.Duration
}

func newSlowLogApplier(config Config) applier {
	return &slowLogApplier{
		slowLogFloor: config.SlowLogFloor,
	}
}

func (s *slowLogApplier) active() bool {
	return s != nil && s.slowLogFloor > 0
}

func (s *slowLogApplier) apply() gin.HandlerFunc {
	if !s.active() {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	syslog.Println("[gin-interceptor] apply slow log:", s.slowLogFloor)

	return func(c *gin.Context) {
		var (
			ctx   = c.Request.Context()
			start = time.Now()
		)

		defer func() {
			if duration := time.Since(start); duration > s.slowLogFloor {
				logger.FromContext(ctx).WithFields(logger.Fields{
					"duration":  duration,
					"url":       c.Request.URL.String(),
					"method":    c.Request.Method,
					"query":     c.Request.URL.Query(),
					"clientIP":  c.ClientIP(),
					"status":    c.Writer.Status(),
					"threshold": s.slowLogFloor,
				}).Warn("slow log")
			}
		}()

		c.Next()
	}
}
