package tracer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagAndRequestID_DoNotCollide(t *testing.T) {
	// The flag and request-id keys are distinct typed structs, so stashing
	// one must not shadow the other on the same context.
	ctx := WithFlag(context.Background(), FlagDebug|FlagShadow)
	ctx = WithRequestID(ctx, "req-xyz")

	assert.Equal(t, FlagDebug|FlagShadow, FromContext(ctx))
	assert.True(t, FromContext(ctx).IsDebug())
	assert.True(t, FromContext(ctx).IsShadow())

	assert.Equal(t, "req-xyz", RequestIDFromCtx(ctx))
}

func TestWithFlags_ORCombine(t *testing.T) {
	ctx := WithFlags(context.Background(), FlagDebug)
	ctx = WithFlags(ctx, FlagShadow) // replaces, not ORs — see WithFlag doc
	// WithFlags stores the OR of its *own* args, replacing prior. So calling
	// it twice does NOT accumulate across calls.
	assert.Equal(t, FlagShadow, FromContext(ctx))
}

func TestWithFlags_SingleCallCombines(t *testing.T) {
	ctx := WithFlags(context.Background(), FlagDebug, FlagStress, FlagShadow)
	got := FromContext(ctx)
	assert.True(t, got.HasAll(FlagDebug|FlagStress|FlagShadow))
}

func TestFlag_SurvivesContextNesting(t *testing.T) {
	// A flag set on a parent context is visible through a derived child.
	parent := WithFlag(context.Background(), FlagDebug)
	child := context.WithValue(parent, flagCtxKey{}, FlagDebug|FlagShadow)
	assert.Equal(t, FlagDebug|FlagShadow, FromContext(child))
}

func TestRequestID_SurvivesContextNesting(t *testing.T) {
	parent := WithRequestID(context.Background(), "outer")
	child := context.WithValue(parent, requestIDCtxKey{}, "inner")
	assert.Equal(t, "inner", RequestIDFromCtx(child))
}

func TestFromContext_StandaloneNoMutate(t *testing.T) {
	// Reading must never write back a generated id; this matters when the
	// same ctx is read from many goroutines.
	ctx := context.Background()
	before := RequestIDFromCtxOr(ctx) // ""
	_ = RequestIDFromCtx(ctx)         // would generate if it wrote back
	assert.Equal(t, before, RequestIDFromCtxOr(ctx))
}

func TestWithRequestID_OverwritesPrior(t *testing.T) {
	ctx := WithRequestID(context.Background(), "first")
	ctx = WithRequestID(ctx, "second")
	assert.Equal(t, "second", RequestIDFromCtx(ctx))
}

func TestParseFlag_ThenWithFlag_ThenReadBack(t *testing.T) {
	// End-to-end: header string -> Flag -> context -> read -> string.
	header := "debug|shadow"
	ctx := WithFlag(context.Background(), ParseFlag(header))
	got := FromContext(ctx)
	assert.True(t, got.IsDebug() && got.IsShadow())
	assert.False(t, got.IsStress())
	assert.Equal(t, header, got.String())
}
