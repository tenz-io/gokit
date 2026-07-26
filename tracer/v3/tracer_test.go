package tracer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlag_BitOps(t *testing.T) {
	assert.Equal(t, FlagNone, Flag(0))
	assert.True(t, FlagDebug.Is(FlagDebug))
	assert.False(t, FlagDebug.Is(FlagStress))

	combo := FlagDebug | FlagStress
	assert.True(t, combo.Is(FlagDebug))
	assert.True(t, combo.Is(FlagStress))
	assert.False(t, combo.Is(FlagShadow))
}

func TestFlag_IsNone(t *testing.T) {
	// FlagNone is not a real flag: Is(FlagNone) is false, unlike v2.
	assert.False(t, (FlagDebug | FlagShadow).Is(FlagNone))
	assert.False(t, FlagNone.Is(FlagNone))
}

func TestFlag_HasAll(t *testing.T) {
	combo := FlagDebug | FlagStress | FlagShadow
	assert.True(t, combo.HasAll(FlagDebug|FlagStress))
	assert.True(t, combo.HasAll(FlagDebug|FlagShadow|FlagStress))
	assert.False(t, combo.HasAll(FlagNone)) // FlagNone is not a flag
}

func TestFlag_ConvenienceMethods(t *testing.T) {
	combo := FlagDebug | FlagShadow
	assert.True(t, combo.IsDebug())
	assert.False(t, combo.IsStress())
	assert.True(t, combo.IsShadow())
}

func TestFromContext(t *testing.T) {
	assert.Equal(t, FlagNone, FromContext(nil))
	assert.Equal(t, FlagNone, FromContext(context.Background()))

	assert.Equal(t, FlagDebug, FromContext(WithFlag(context.Background(), FlagDebug)))

	combo := FromContext(WithFlags(context.Background(), FlagDebug, FlagStress))
	assert.True(t, combo.Is(FlagDebug) && combo.Is(FlagStress))
	assert.False(t, combo.Is(FlagShadow))
}

func TestWithFlag_Overwrites(t *testing.T) {
	// WithFlag replaces; it does not OR-combine with the prior value.
	ctx := WithFlag(context.Background(), FlagDebug)
	ctx = WithFlag(ctx, FlagShadow)
	assert.Equal(t, FlagShadow, FromContext(ctx))
	assert.False(t, FromContext(ctx).IsDebug())
}

func TestWithFlag_NilCtx(t *testing.T) {
	// A nil ctx is returned unchanged (nil) rather than panicking.
	assert.Nil(t, WithFlag(nil, FlagDebug))
	assert.Nil(t, WithFlags(nil, FlagDebug, FlagShadow))
	assert.Equal(t, FlagNone, FromContext(nil))
}

func TestRequestIDFromCtx_AutoGenerate(t *testing.T) {
	// Unset context -> a fresh non-empty ID is generated.
	id := RequestIDFromCtx(context.Background())
	assert.NotEmpty(t, id)
	assert.Len(t, id, 32) // uuid v4 minus 4 dashes = 32 hex chars
	assert.NotContains(t, id, "-")

	// nil ctx also yields a generated ID, not a panic.
	assert.NotEmpty(t, RequestIDFromCtx(nil))
}

func TestRequestIDFromCtx_RoundTrip(t *testing.T) {
	ctx := WithRequestID(context.Background(), "req-123")
	assert.Equal(t, "req-123", RequestIDFromCtx(ctx))
}

func TestRequestIDFromCtx_KeepsExisting(t *testing.T) {
	// A caller-set ID is never replaced by a generated one.
	ctx := WithRequestID(context.Background(), "caller-id")
	assert.Equal(t, "caller-id", RequestIDFromCtx(ctx))
}

func TestRequestIDFromCtxOr(t *testing.T) {
	assert.Empty(t, RequestIDFromCtxOr(nil))
	assert.Empty(t, RequestIDFromCtxOr(context.Background()))

	ctx := WithRequestID(context.Background(), "abc")
	assert.Equal(t, "abc", RequestIDFromCtxOr(ctx))
}

func TestWithRequestID_EmptyNoOp(t *testing.T) {
	// Stashing an empty id is a no-op: nothing is stored on the context.
	ctx := WithRequestID(context.Background(), "")
	assert.Empty(t, RequestIDFromCtxOr(ctx)) // nothing stored -> ""
	assert.Len(t, RequestIDFromCtx(ctx), 32) // auto-generate path still yields a fresh id
}

func TestWithRequestID_NilCtx(t *testing.T) {
	// nil ctx short-circuits before touching context.WithValue.
	assert.Nil(t, WithRequestID(nil, "id"))
}

func TestNewRequestID(t *testing.T) {
	a, b := NewRequestID(), NewRequestID()
	assert.Len(t, a, 32)
	assert.Len(t, b, 32)
	assert.NotEqual(t, a, b) // randomness sanity check
	assert.NotContains(t, a, "-")
}
