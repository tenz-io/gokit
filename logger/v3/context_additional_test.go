package logger

import (
	"context"
	"testing"
)

func TestContextPreservesAttachedEntry(t *testing.T) {
	e := NewEntry(Config{Level: DebugLevel, Console: false})
	ctx := WithLogger(context.Background(), e)
	if got := FromContext(ctx); got != e {
		t.Fatalf("FromContext() = %p, want %p", got, e)
	}
	dst := context.WithValue(context.Background(), struct{ name string }{"dst"}, true)
	copied := CopyToContext(ctx, dst)
	if got := FromContext(copied); got != e {
		t.Fatalf("copied entry = %p, want %p", got, e)
	}
}

func TestCopyToContextWithoutAttachedEntryIsNoOp(t *testing.T) {
	dst := context.Background()
	if got := CopyToContext(context.Background(), dst); got != dst {
		t.Fatal("CopyToContext should not attach the global fallback")
	}
	if got := CopyToContext(nil, dst); got != dst {
		t.Fatal("nil source should return destination unchanged")
	}
	if got := CopyToContext(context.Background(), nil); got != nil {
		t.Fatal("nil destination should remain nil")
	}
}

func TestContextRejectsTypedNilEntry(t *testing.T) {
	var e *logEntry
	ctx := context.Background()
	if got := WithLogger(ctx, e); got != ctx {
		t.Fatal("typed nil entry should be a no-op")
	}
	ctx = context.WithValue(ctx, ctxKey{}, Entry(e))
	if got := FromContext(ctx); got != L() {
		t.Fatal("typed nil context value should fall back to global logger")
	}
}

func TestWithLoggerNilContext(t *testing.T) {
	if got := WithLogger(nil, L()); got != nil {
		t.Fatal("nil context should remain nil")
	}
}
