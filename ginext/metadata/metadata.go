package metadata

import (
	"context"

	"github.com/gin-gonic/gin"
)

type mdCtxKey struct{}

type MD struct {
	Header   map[string]string
	Path     string
	Method   string
	Host     string
	ClientIP string

	// Additional fields

	// ContentType from header Content-Type
	ContentType string
	// Authorization from header Authorization
	Authorization string
	// RequestID from header X-Request-ID
	RequestID string
	// SessionID from header X-Session-ID
	SessionID string
	// RequestMode from header X-Request-Mode
	// modes: normal, debug, shadow
	RequestMode string
}

func New(c *gin.Context) *MD {
	md := &MD{
		Path:     c.Request.URL.Path,
		Method:   c.Request.Method,
		Host:     c.Request.Host,
		ClientIP: c.ClientIP(),
		Header:   make(map[string]string, len(c.Request.Header)),
	}
	for k, val := range c.Request.Header {
		md.setHeader(k, val...)
	}

	md.additional(c)

	return md
}

func (md *MD) additional(c *gin.Context) {
	// ContentType from header Content-Type
	md.ContentType = c.ContentType()

	// Authorization from header Authorization
	md.Authorization = c.GetHeader("Authorization")

	// RequestID from header X-Request-ID
	md.RequestID = c.GetHeader("X-Request-ID")

	// SessionID from header X-Session-ID
	md.SessionID = c.GetHeader("X-Session-ID")

	// RequestMode from header X-Request-Mode
	md.RequestMode = c.GetHeader("X-Request-Mode")
}

func (md *MD) setHeader(k string, vals ...string) {
	if len(vals) == 0 {
		return
	}
	md.Header[k] = vals[0]
}

func WithMetadata(ctx context.Context, md *MD) context.Context {
	return context.WithValue(ctx, mdCtxKey{}, md)
}

func FromContext(ctx context.Context) (*MD, bool) {
	md, ok := ctx.Value(mdCtxKey{}).(*MD)
	if !ok {
		return nil, false
	}

	return md, true
}
