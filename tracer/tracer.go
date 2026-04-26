// Package tracer provides context-based request ID propagation and debug/stress/shadow flag management.
package tracer

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// Flag is a bitmask for traffic mode flags.
type Flag int

const (
	FlagNone Flag = 0
	// FlagDebug enables debug mode.
	FlagDebug Flag = 1 << iota
	// FlagStress enables stress test mode.
	FlagStress
	// FlagShadow enables shadow mode (traffic dumping and replaying).
	FlagShadow
)

// Is reports whether all bits in flagToCheck are set in f.
func (f Flag) Is(flagToCheck Flag) bool { return f&flagToCheck == flagToCheck }

// HasAll reports whether all specified flags are set.
func (f Flag) HasAll(flags Flag) bool { return f&flags == flags }

// IsDebug reports whether FlagDebug is set.
func (f Flag) IsDebug() bool { return f.Is(FlagDebug) }

// IsStress reports whether FlagStress is set.
func (f Flag) IsStress() bool { return f.Is(FlagStress) }

// IsShadow reports whether FlagShadow is set.
func (f Flag) IsShadow() bool { return f.Is(FlagShadow) }

// Context keys are unexported to prevent collisions.
type (
	flagCtxKey      string
	requestIDCtxKey string
)

const (
	flagKey      flagCtxKey      = "_flag_ctx_key"
	requestIDKey requestIDCtxKey = "_requestId_ctx_key"
)

// FromContext returns the trace flag from the context, or FlagNone if not set.
func FromContext(ctx context.Context) Flag {
	if ctx == nil {
		return FlagNone
	}
	if f, ok := ctx.Value(flagKey).(Flag); ok {
		return f
	}
	return FlagNone
}

// WithFlag returns a new context with the given flag set.
func WithFlag(ctx context.Context, flag Flag) context.Context {
	return context.WithValue(ctx, flagKey, flag)
}

// WithFlags combines multiple flags and sets them in the context.
func WithFlags(ctx context.Context, flags ...Flag) context.Context {
	var f Flag
	for _, fl := range flags {
		f |= fl
	}
	return WithFlag(ctx, f)
}

// RequestIdFromCtx returns the request ID from the context.
// If none is set, a new UUID-based ID is generated and returned.
func RequestIdFromCtx(ctx context.Context) string {
	if ctx != nil {
		if id, ok := ctx.Value(requestIDKey).(string); ok && id != "" {
			return id
		}
	}
	return newRequestID()
}

// RequestIdFromCtxOr returns the request ID from the context, or empty string if not set.
func RequestIdFromCtxOr(ctx context.Context) string {
	if ctx != nil {
		if id, ok := ctx.Value(requestIDKey).(string); ok {
			return id
		}
	}
	return ""
}

// WithRequestId sets the request ID in the context.
func WithRequestId(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func newRequestID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}
