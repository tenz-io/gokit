package monitor

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// These benchmarks exercise the Begin/End hot path and the raw point methods to
// confirm the WithLabelValues-based implementation allocates nothing on the
// steady-state path (a pre-existing series, already interned in the vector's
// metricMap). Compare to the v2 style — labelsOf(map) + With(labels) — which
// allocated a Labels map on every call.
//
// Run: go test -bench=. -benchmem

// benchmarkExporter builds a configured Exporter against an isolated Registry
// and pre-warms the label series so the steady-state (no new series) path is
// measured.
func benchmarkExporter(b *testing.B) Exporter {
	b.Helper()
	reg := prometheus.NewRegistry()
	resetForTest()
	if err := Configure(WithRegistry(reg)); err != nil {
		b.Fatalf("configure: %v", err)
	}
	exp := NewExporter("svc")
	// Pre-create the series so the benchmark measures steady-state, not
	// first-touch metric creation.
	ctx := context.Background()
	exp.Incr(ctx, "op", codeOK, optActive)
	exp.Observe(ctx, "op", codeOK, 1.0)
	exp.Count(ctx, "op", codeOK, valNA)
	exp.Sample(ctx, "op", codeOK, 1.0, valNA)
	return exp
}

func BenchmarkExporterPointMethods(b *testing.B) {
	exp := benchmarkExporter(b)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// One round of each point method per iteration.
		exp.Incr(ctx, "op", codeOK, optActive)
		exp.Observe(ctx, "op", codeOK, 1.0)
		exp.Count(ctx, "op", codeOK, valNA)
		exp.Sample(ctx, "op", codeOK, 1.0, valNA)
		exp.Decr(ctx, "op", codeOK, optActive)
	}
}

// BenchmarkBeginEnd measures the full single-flight recorder path, which is
// what every instrumented call site pays. The 2 allocs/op are the Recorder
// struct (1) and the context.WithoutCancel ctx captured for End (1) — the
// latter preserves trace/tenant values for custom Exporters after the request
// ctx is cancelled. The built-in Exporter ignores ctx, so a custom Exporter
// is the only consumer of that second allocation.
func BenchmarkBeginEnd(b *testing.B) {
	exp := benchmarkExporter(b)
	ctx := WithExporter(context.Background(), exp)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := Begin(ctx, "op")
		rec.End()
	}
}
