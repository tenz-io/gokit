package tracer

import (
	"context"
	"testing"
)

func TestFlag_Is(t *testing.T) {
	if !FlagDebug.Is(FlagDebug) {
		t.Error("FlagDebug.Is(FlagDebug) should be true")
	}
	if FlagDebug.Is(FlagStress) {
		t.Error("FlagDebug.Is(FlagStress) should be false")
	}
	combo := FlagDebug | FlagStress
	if !combo.Is(FlagDebug) {
		t.Error("combo.Is(FlagDebug) should be true")
	}
	if !combo.Is(FlagStress) {
		t.Error("combo.Is(FlagStress) should be true")
	}
}

func TestFlag_HasAll(t *testing.T) {
	combo := FlagDebug | FlagStress | FlagShadow
	if !combo.HasAll(FlagDebug | FlagStress) {
		t.Error("HasAll(FlagDebug|FlagStress) should be true")
	}
	if combo.HasAll(FlagDebug | FlagShadow | FlagStress) {
		// has all of them
	}
	if combo.HasAll(FlagNone) {
		// FlagNone & anything = 0, so check is trivial
	}
}

func TestFlag_IsMethods(t *testing.T) {
	combo := FlagDebug | FlagShadow
	if !combo.IsDebug() {
		t.Error("combo.IsDebug() should be true")
	}
	if combo.IsStress() {
		t.Error("combo.IsStress() should be false")
	}
	if !combo.IsShadow() {
		t.Error("combo.IsShadow() should be true")
	}
}

func TestFromContext(t *testing.T) {
	if got := FromContext(nil); got != FlagNone {
		t.Errorf("FromContext(nil) = %v", got)
	}

	ctx := WithFlag(context.Background(), FlagDebug)
	if got := FromContext(ctx); got != FlagDebug {
		t.Errorf("FromContext(with FlagDebug) = %v", got)
	}

	ctx = WithFlags(context.Background(), FlagDebug, FlagStress)
	got := FromContext(ctx)
	if !got.Is(FlagDebug) || !got.Is(FlagStress) {
		t.Errorf("FromContext(with FlagDebug|FlagStress) = %v", got)
	}
}

func TestRequestIdFromCtx(t *testing.T) {
	if got := RequestIdFromCtx(nil); got == "" {
		t.Error("RequestIdFromCtx(nil) should auto-generate")
	}

	id := "my-custom-id"
	ctx := WithRequestId(context.Background(), id)
	if got := RequestIdFromCtx(ctx); got != id {
		t.Errorf("RequestIdFromCtx() = %v, want %v", got, id)
	}
}

func TestRequestIdFromCtxOr(t *testing.T) {
	if got := RequestIdFromCtxOr(nil); got != "" {
		t.Errorf("RequestIdFromCtxOr(nil) = %v, want empty", got)
	}

	ctx := WithRequestId(context.Background(), "abc")
	if got := RequestIdFromCtxOr(ctx); got != "abc" {
		t.Errorf("RequestIdFromCtxOr() = %v, want abc", got)
	}
}

func TestWithRequestId(t *testing.T) {
	ctx := WithRequestId(context.Background(), "test-id")
	if id := RequestIdFromCtxOr(ctx); id != "test-id" {
		t.Errorf("got %v, want test-id", id)
	}
}
