package ginterceptor

import (
	"context"
	syslog "log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/monitor"
)

const (
	requestIdCtxKey = requestIdCtxKeyType("requestId_ctx_key")
)

type requestIdCtxKeyType string

func (i *interceptor) ApplyTracking() gin.HandlerFunc {
	syslog.Println("[httpgin] apply tracking")

	return func(c *gin.Context) {
		var (
			url = c.Request.URL.Path
			ctx = c.Request.Context()
		)

		// metrics tracking
		if i.config.EnableMetrics {
			ctx = monitor.InitSingleFlight(ctx, url)
		}

		requestId := RequestIdFromCtx(ctx)
		ctx = WithRequestId(ctx, requestId)

		le := logger.WithFields(logger.Fields{
			"url": url,
		}).WithTracing(requestId)
		ctx = logger.WithLogger(ctx, le)

		te := logger.WithTrafficTracing(ctx, requestId).
			WithFields(logger.Fields{
				"url": url,
			}).
			WithIgnores(
				"password",
				//"Authorization",
			)
		ctx = logger.WithTrafficEntry(ctx, te)

		WithContext(c, ctx)

		defer func() {
			c.Writer.Header().Set("X-Request-Id", requestId)
		}()

		c.Next()
	}
}

// RequestIdFromCtx returns the value associated with this context for key, or nil
func RequestIdFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if requestId, ok := ctx.Value(requestIdCtxKey).(string); ok {
		return requestId
	}

	return newRequestId()
}

// WithRequestId returns a copy of parent in which the value associated with key is val.
func WithRequestId(ctx context.Context, requestID string) context.Context {
	ctx = context.WithValue(ctx, requestIdCtxKey, requestID)
	return ctx
}

func newRequestId() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

// WithContext returns a copy of parent in which the value associated with key is val.
func WithContext(c *gin.Context, ctx context.Context) {
	if c == nil {
		return
	}
	c.Request = c.Request.WithContext(ctx)
}
