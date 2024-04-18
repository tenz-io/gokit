package tracer

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

var (
	requestIdCtxKey = requestIdCtxKeyType("_requestId_ctx_key")
)

type requestIdCtxKeyType string

// Flag is a type to represent the trace flag of traffic.
type Flag int

const (
	FlagNone Flag = 0
)

const (
	// FlagDebug is a flag to enable debug mode.
	FlagDebug Flag = 1 << iota // 1
	// FlagStress is a flag to enable stress test mode.
	FlagStress
	// FlagShadow is a flag to enable shadow mode.
	// used for traffic dumping and replaying.
	FlagShadow
)

type flagCtxKey string

const (
	flagCtxKeyFlag = flagCtxKey("_flag_ctx_key_flag")
)

// FromContext returns the trace flag from the context.
func FromContext(ctx context.Context) (flag Flag) {
	if ctx == nil {
		return FlagNone
	}
	if f, ok := ctx.Value(flagCtxKeyFlag).(Flag); ok {
		return f
	}
	return FlagNone
}

// WithFlag returns a new context with the trace flag.
// flag can be combined with bitwise OR.
// e.g. WithFlag(ctx, FlagDebug), WithFlag(ctx, FlagDebug|FlagStress), etc.
func WithFlag(ctx context.Context, flag Flag) context.Context {
	return context.WithValue(ctx, flagCtxKeyFlag, flag)
}

// WithFlags returns a new context with the trace flag.
func WithFlags(ctx context.Context, flags ...Flag) context.Context {
	var flag = FlagNone
	for _, f := range flags {
		flag |= f
	}
	return WithFlag(ctx, flag)
}

// Is returns true if the trace flag is set.
func (f Flag) Is(flagToCheck Flag) bool {
	return f&flagToCheck == flagToCheck
}

// IsDebug returns true if the trace flag is set.
func (f Flag) IsDebug() bool {
	return f.Is(FlagDebug)
}

// IsStress returns true if the trace flag is set.
func (f Flag) IsStress() bool {
	return f.Is(FlagStress)
}

// IsShadow returns true if the trace flag is set.
func (f Flag) IsShadow() bool {
	return f.Is(FlagShadow)
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
