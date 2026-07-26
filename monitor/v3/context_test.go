package monitor

import (
	"context"
	"testing"
)

func TestFromContextNoExporterReturnsNoop(t *testing.T) {
	// Un-instrumented context: FromContext returns a non-nil no-op Exporter.
	exp := FromContext(context.Background())
	if exp == nil {
		t.Fatal("FromContext returned nil; want non-nil noop Exporter")
	}
	if !isNoop(exp) {
		t.Errorf("FromContext on bare ctx returned real Exporter; want noop")
	}
}

func TestHasExporter(t *testing.T) {
	ctx := context.Background()
	if HasExporter(ctx) {
		t.Error("HasExporter = true before Init; want false")
	}
	ctx = Init(ctx, "svc")
	if !HasExporter(ctx) {
		t.Error("HasExporter = false after Init; want true")
	}
}

func TestInitCreatesThenReuses(t *testing.T) {
	freshRegistry(t)
	ctx := context.Background()

	ctx = Init(ctx, "svc")
	first := exporterFromContext(ctx)
	if first == nil {
		t.Fatal("Init did not inject an Exporter")
	}
	if got := first.Cmd(); got != "svc" {
		t.Errorf("first Init cmd = %q, want svc", got)
	}

	// Second Init must reuse the existing flight, not overwrite with a new one.
	ctx2 := Init(ctx, "other")
	second := exporterFromContext(ctx2)
	if second != first {
		t.Error("second Init overwrote the existing Exporter; want reuse")
	}
	if got := second.Cmd(); got != "svc" {
		t.Errorf("reused Exporter cmd = %q, want svc (unchanged)", got)
	}
}

func TestInitEmptyCmdNormalizes(t *testing.T) {
	freshRegistry(t)
	ctx := Init(context.Background(), "")
	exp := exporterFromContext(ctx)
	if got := exp.Cmd(); got != valNA {
		t.Errorf("Init(\"\") cmd = %q, want %q", got, valNA)
	}
}

func TestInitNilContext(t *testing.T) {
	freshRegistry(t)
	// Init with a nil ctx must not panic; it falls back to Background.
	ctx := Init(nil, "svc")
	if !HasExporter(ctx) {
		t.Fatal("Init(nil, ...) did not inject an Exporter")
	}
}

func TestFromContextNilContext(t *testing.T) {
	// FromContext with nil must return the noop, not panic.
	exp := FromContext(nil)
	if exp == nil {
		t.Fatal("FromContext(nil) returned nil; want noop")
	}
}

func TestCopyToContextForwardsExporter(t *testing.T) {
	freshRegistry(t)
	src := Init(context.Background(), "svc")
	dst := context.Background()

	if HasExporter(dst) {
		t.Fatal("dst already has an Exporter before CopyToContext")
	}
	out := CopyToContext(src, dst)
	if !HasExporter(out) {
		t.Fatal("CopyToContext did not forward the Exporter to dst")
	}
	// Forwarded Exporter must be the same instance.
	if exporterFromContext(out) != exporterFromContext(src) {
		t.Error("CopyToContext forwarded a different Exporter instance")
	}
}

func TestCopyToContextNoExporterLeavesDstUnchanged(t *testing.T) {
	dst := context.Background()
	out := CopyToContext(context.Background(), dst)
	if HasExporter(out) {
		t.Error("CopyToContext from a ctx with no Exporter injected one into dst")
	}
}

func TestCopyToContextNilArgs(t *testing.T) {
	if out := CopyToContext(nil, context.Background()); HasExporter(out) {
		t.Error("CopyToContext(nil, bg) should not inject anything")
	}
	if out := CopyToContext(context.Background(), nil); out != nil {
		t.Error("CopyToContext(bg, nil) should return nil dst unchanged")
	}
}

func TestWithExporter(t *testing.T) {
	freshRegistry(t)
	exp := NewExporter("svc")
	ctx := WithExporter(context.Background(), exp)
	if exporterFromContext(ctx) != exp {
		t.Error("WithExporter did not inject the given Exporter")
	}
	if !HasExporter(ctx) {
		t.Error("HasExporter false after WithExporter")
	}
}

// isNoop reports whether exp is the package's noop Exporter.
func isNoop(exp Exporter) bool {
	_, ok := exp.(*noopExporter)
	return ok
}
