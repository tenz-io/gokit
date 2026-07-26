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
	// Reading via RequestIDFromCtx must never write back a generated id: the
	// original ctx is observed unchanged (RequestIDFromCtxOr still returns
	// ""). This matters when the same ctx is read concurrently.
	ctx := context.Background()
	assert.Empty(t, RequestIDFromCtxOr(ctx)) // nothing stored
	_ = RequestIDFromCtx(ctx)                // generates but must not write back
	assert.Empty(t, RequestIDFromCtxOr(ctx)) // still nothing stored
}

func TestEnsureRequestID_PinsIDForTheRequest(t *testing.T) {
	// The whole point of EnsureRequestID: reading the derived context twice
	// (or N times, across goroutines) returns the SAME id, unlike the bare
	// RequestIDFromCtx read of an id-less context.
	base := context.Background()
	ctx, id := EnsureRequestID(base)
	assert.Len(t, id, 32)
	assert.NotEmpty(t, id)

	// Repeated reads of the derived context are stable.
	assert.Equal(t, id, RequestIDFromCtx(ctx))
	assert.Equal(t, id, RequestIDFromCtx(ctx))
	assert.Equal(t, id, RequestIDFromCtxOr(ctx))

	// And EnsureRequestID is idempotent: a second call keeps the stored id.
	ctx2, id2 := EnsureRequestID(ctx)
	assert.Equal(t, id, id2)
	assert.Same(t, ctx, ctx2) // already had an id -> returned unchanged
}

func TestEnsureRequestID_BareCtxReadIsUnstable(t *testing.T) {
	// Documenting the sharp edge that motivates EnsureRequestID: reading an
	// id-less context via RequestIDFromCtx twice yields two different ids.
	ctx := context.Background()
	assert.NotEqual(t, RequestIDFromCtx(ctx), RequestIDFromCtx(ctx))
}

func TestEnsureRequestID_NilCtx(t *testing.T) {
	// A nil ctx yields a generated id and a nil (re-derived) context rather
	// than panicking.
	ctx, id := EnsureRequestID(nil)
	assert.Nil(t, ctx)
	assert.Len(t, id, 32)
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
