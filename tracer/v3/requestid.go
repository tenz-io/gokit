package tracer

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// requestIDCtxKey is the typed context key under which a request ID string
// is stashed. Typed (not a string) so it cannot collide with keys other
// packages attach to the same context.
type requestIDCtxKey struct{}

// WithRequestID returns a derived context carrying id as the request ID.
// Stashing an empty id is a no-op (the context is returned unchanged) so a
// caller can chain WithRequestID(ctx, maybeEmpty) without guarding. A nil
// ctx is returned unchanged.
func WithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil || id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDCtxKey{}, id)
}

// RequestIDFromCtx returns the request ID stored on ctx by [WithRequestID].
// When none is stored (or ctx is nil) it auto-generates a fresh ID via
// [NewRequestID] so downstream never observes an empty ID — use this on the
// read path of inbound handling.
func RequestIDFromCtx(ctx context.Context) string {
	if ctx != nil {
		if id, ok := ctx.Value(requestIDCtxKey{}).(string); ok && id != "" {
			return id
		}
	}
	return NewRequestID()
}

// RequestIDFromCtxOr returns the request ID stored on ctx, or "" when none
// is stored. Unlike [RequestIDFromCtx] it does not auto-generate, which suits
// "only forward if already present" checks.
func RequestIDFromCtxOr(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return id
	}
	return ""
}

// NewRequestID generates a fresh, dash-stripped UUIDv4 string (32 hex chars).
// It is the fallback used by [RequestIDFromCtx] and is exported so callers
// that build an ID before there is a context can reuse the same generator.
func NewRequestID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
