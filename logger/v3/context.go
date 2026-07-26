package logger

import (
	"context"
	"reflect"
)

// ctxKey is the unexported context key under which an Entry is stashed.
type ctxKey struct{}

// FromContext returns the Entry attached to ctx, or the global logger when
// none is attached (including when ctx is nil). This is the read side of the
// context-propagation chain: callers stash a per-request Entry (carrying
// request_id, url, ...) onto ctx with WithLogger, then any function down the
// call chain recovers it with FromContext so the shared fields are stamped
// onto every log line without re-passing them.
func FromContext(ctx context.Context) Entry {
	if ctx == nil {
		return current()
	}
	if e, ok := ctx.Value(ctxKey{}).(Entry); ok && !isNilEntry(e) {
		return e
	}
	return current()
}

// WithLogger attaches an Entry to ctx, returning a derived context. Passing a
// nil Entry is a no-op so callers can chain WithLogger(ctx, logger.With...)
// without guarding.
func WithLogger(ctx context.Context, e Entry) context.Context {
	if ctx == nil || isNilEntry(e) {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, e)
}

// CopyToContext copies an explicitly attached Entry from srcCtx onto dstCtx,
// returning the derived dstCtx. When srcCtx has no attached Entry, dstCtx is
// returned unchanged; the global fallback is not captured. This is useful
// when crossing a context boundary (e.g. spawning a goroutine or entering a
// transport that builds a fresh context) while preserving a request logger.
func CopyToContext(srcCtx, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}
	e, ok := srcCtx.Value(ctxKey{}).(Entry)
	if !ok || isNilEntry(e) {
		return dstCtx
	}
	return WithLogger(dstCtx, e)
}

func isNilEntry(e Entry) bool {
	if e == nil {
		return true
	}
	v := reflect.ValueOf(e)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
