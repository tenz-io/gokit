package ginext

import (
	"context"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/logger"
	"github.com/tenz-io/gokit/tracer"
)

const (
	headerNameRequestId   = "X-Request-Id"
	headerNameRequestFlag = "X-Request-Flag"
)

const (
	requestIdCtxKey = requestIdCtxKeyType("_requestId_ctx_key")
)

type requestIdCtxKeyType string

// deprecated
// use RpcInterceptor instead
type trackingApplier struct {
	enable bool
}

func newTrackingApplier(config Config) applier {
	return &trackingApplier{
		enable: config.EnableTracking,
	}

}

func (t *trackingApplier) active() bool {
	return t != nil && t.enable
}

func (t *trackingApplier) apply() gin.HandlerFunc {
	if !t.active() {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	log.Println("[gin-interceptor] apply tracking")

	return func(c *gin.Context) {
		var (
			url   = c.Request.URL.Path
			reqID = requestIdFromGinCtx(c)
			flag  = requestFlagFromHeader(c)
			ctx   = tracer.WithRequestId(c.Request.Context(), reqID)
		)

		// inject trace flag into context
		ctx = tracer.WithFlags(ctx, flag)

		// inject logger into context
		ctx = logger.WithLogger(
			ctx,
			logger.WithTracing(reqID).
				WithFields(logger.Fields{
					"url": url,
				}),
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

	return tracer.RequestIdFromCtx(c.Request.Context())
}

func requestFlagFromHeader(c *gin.Context) tracer.Flag {
	if c == nil {
		return tracer.FlagNone
	}
	headFlag := c.GetHeader(headerNameRequestFlag)
	if headFlag == "" {
		return tracer.FlagNone
	}

	// convert headFlag from string to int
	flag, err := strconv.Atoi(headFlag)
	if err != nil {
		return tracer.FlagNone
	}

	return tracer.Flag(flag)
}

// WithContext returns a copy of parent in which the value associated with key is val.
func WithContext(c *gin.Context, ctx context.Context) {
	if c == nil {
		return
	}
	c.Request = c.Request.WithContext(ctx)
}
