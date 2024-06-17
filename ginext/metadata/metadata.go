package metadata

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/tracer"
)

const (
	headerAuthorization = "Authorization"
	headerRequestID     = "X-Request-Id"
	headerSessionID     = "X-Session-Id"
	headerRequestFlag   = "X-Request-Flag"
)

const (
	usernameKey = "username"
	roleKey     = "role"
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
	// RequestFlag from header X-Request-Flag
	// modes: normal, debug, shadow, stress
	// or combine of modes: debug|shadow
	RequestFlag string
	// Cmd grpc command/service
	Cmd string

	Username string
	Role     string
}

func New(c *gin.Context, cmd string) *MD {
	md := &MD{
		Path:     c.Request.URL.Path,
		Method:   c.Request.Method,
		Host:     c.Request.Host,
		ClientIP: c.ClientIP(),
		Header:   make(map[string]string, len(c.Request.Header)),
		Cmd:      cmd,
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
	md.Authorization = c.GetHeader(headerAuthorization)

	// RequestID from header X-Request-ID
	md.RequestID = c.GetHeader(headerRequestID)
	if md.RequestID == "" {
		md.RequestID = tracer.RequestIdFromCtx(c.Request.Context())
	}
	c.Request.WithContext(tracer.WithRequestId(c.Request.Context(), md.RequestID))
	c.Writer.Header().Set(headerRequestID, md.RequestID)

	// SessionID from header X-Session-ID
	md.SessionID = c.GetHeader(headerSessionID)

	// RequestMode from header X-Request-Mode
	md.RequestFlag = c.GetHeader(headerRequestFlag)

	// Username from context
	md.Username = c.GetString(usernameKey)
	md.Role = c.GetString(roleKey)
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
