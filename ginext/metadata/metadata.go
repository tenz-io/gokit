package metadata

import (
	"context"
	"github.com/gin-gonic/gin"
)

type mdCtxKey struct{}

type MD struct {
	Header   map[string][]string
	Path     string
	Method   string
	Host     string
	ClientIP string
}

func New(c *gin.Context) *MD {
	md := &MD{
		Path:     c.Request.URL.Path,
		Method:   c.Request.Method,
		Host:     c.Request.Host,
		ClientIP: c.ClientIP(),
		Header:   make(map[string][]string, len(c.Request.Header)),
	}
	for k, val := range c.Request.Header {
		md.setHeader(k, val...)
	}

	return md
}

func (md *MD) setHeader(k string, vals ...string) {
	if len(vals) == 0 {
		return
	}
	md.Header[k] = vals
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
