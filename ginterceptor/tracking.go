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
	headerNameRequestId = "X-Request-Id"
)

const (
	requestIdCtxKey = requestIdCtxKeyType("requestId_ctx_key")
)

type requestIdCtxKeyType string

func (i *interceptor) ApplyTracking() gin.HandlerFunc {
	syslog.Println("[gin-interceptor] apply tracking")

	return func(c *gin.Context) {
		var (
			url   = c.Request.URL.Path
			reqID = requestIdFromGinCtx(c)
			ctx   = WithRequestId(c.Request.Context(), reqID)
		)

		// metrics tracking
		if i.config.EnableMetrics {
			ctx = monitor.InitSingleFlight(ctx, url)
		}

		// inject logger into context
		ctx = logger.WithLogger(
			ctx,
			logger.WithTracing(reqID).
				WithFields(logger.Fields{
					"url": url,
				}),
		)

		// inject traffic logger into context
		ctx = logger.WithTrafficEntry(
			ctx,
			logger.WithTrafficTracing(ctx, reqID).
				WithFields(logger.Fields{
					"url": url,
				}).WithIgnores(
				"password",
				//"Authorization",
			),
		)

		// update gin context
		WithContext(c, ctx)

		defer func() {
			c.Writer.Header().Set(headerNameRequestId, reqID)
		}()

		c.Next()
	}
}

func requestIdFromGinCtx(c *gin.Context) string {
	if c == nil {
		return ""
	}

	if requestId := c.GetHeader(headerNameRequestId); requestId != "" {
		return requestId
	}

	return RequestIdFromCtx(c.Request.Context())

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
