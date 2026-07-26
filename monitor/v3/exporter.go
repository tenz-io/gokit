package monitor

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is the cmd-scoped metrics exporter interface. One Exporter per
// entry command (created/reused by Init and injected into the context) is
// shared by every downstream Begin/End along the call chain — this is the
// single-flight injection model. All point methods are synchronous and
// allocation-light: they normalize label values (empty opt → NA, non-zero
// code → err) and call WithLabelValues, the Prometheus slice fast path, so no
// map is allocated on the hot path.
//
// Exporter is an interface so FromContext can return a non-nil no-op
// implementation when nothing was injected, letting callers skip nil-checks.
type Exporter interface {
	// Cmd returns the bound cmd label value.
	Cmd() string

	// Set / Incr / Decr operate on the gauge (instantaneous values, active counts).
	Set(ctx context.Context, dsCmd, code string, val float64, opt string)
	Incr(ctx context.Context, dsCmd, code, opt string)
	Decr(ctx context.Context, dsCmd, code, opt string)
	// Count / CountDelta operate on the counter (cumulative totals).
	Count(ctx context.Context, dsCmd, code, opt string)
	// CountDelta adds delta to the counter. delta is uint64 because a
	// Prometheus counter is monotonically non-decreasing: a negative delta
	// would panic inside prometheus.Counter.Add. The unsigned type makes that
	// contract unrepresentable at the call site.
	CountDelta(ctx context.Context, dsCmd, code string, delta uint64, opt string)
	// Observe records millis into the latency histogram; code is normalized to ok/err.
	Observe(ctx context.Context, dsCmd, code string, millis float64)
	// Sample records a value into the data-size summary; code is normalized to ok/err.
	Sample(ctx context.Context, dsCmd, code string, val float64, opt string)
}

// exporter is the concrete Prometheus-backed Exporter. It binds one cmd and
// resolves the four metric families once at construction.
type exporter struct {
	cmd       string
	counter   *prometheus.CounterVec
	gauge     *prometheus.GaugeVec
	histogram *prometheus.HistogramVec
	summary   *prometheus.SummaryVec
}

// NewExporter builds an Exporter bound to cmd. An empty cmd is normalized to
// NA so the cmd label is always set. The metric families are resolved once
// (registering them against the active Registerer on first use).
//
// NewExporter does not require Init to have been called: if it hasn't, the
// package falls back to prometheus.DefaultRegisterer.
func NewExporter(cmd string) Exporter {
	if cmd == "" {
		cmd = valNA
	}
	counter, gauge, histogram, summary := snapshotMetrics()
	return &exporter{
		cmd:       cmd,
		counter:   counter,
		gauge:     gauge,
		histogram: histogram,
		summary:   summary,
	}
}

func (e *exporter) Cmd() string { return e.cmd }

// Set the gauge to an absolute value.
func (e *exporter) Set(ctx context.Context, dsCmd, code string, val float64, opt string) {
	e.gauge.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Set(val)
}

// Incr the gauge by 1.
func (e *exporter) Incr(ctx context.Context, dsCmd, code, opt string) {
	e.gauge.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Inc()
}

// Decr the gauge by 1.
func (e *exporter) Decr(ctx context.Context, dsCmd, code, opt string) {
	e.gauge.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Dec()
}

// Count increments the counter by 1.
func (e *exporter) Count(ctx context.Context, dsCmd, code, opt string) {
	e.counter.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Inc()
}

// CountDelta adds delta to the counter.
func (e *exporter) CountDelta(ctx context.Context, dsCmd, code string, delta uint64, opt string) {
	e.counter.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Add(float64(delta))
}

// Observe records millis into the latency histogram. The code is normalized to
// ok/err so latency cardinality stays bounded regardless of how callers spell
// their result codes.
func (e *exporter) Observe(ctx context.Context, dsCmd, code string, millis float64) {
	e.histogram.WithLabelValues(e.cmd, dsCmd, normalizeCode(code), valNA).Observe(millis)
}

// Sample records val into the data-size summary. The code is normalized to
// ok/err for the same cardinality reason as Observe.
func (e *exporter) Sample(ctx context.Context, dsCmd, code string, val float64, opt string) {
	e.summary.WithLabelValues(e.cmd, dsCmd, normalizeCode(code), normalizeOpt(opt)).Observe(val)
}

// noopExporter is a non-nil Exporter returned by FromContext when no Exporter
// was injected into the context. Callers can use it without nil-checking;
// every method is a no-op so disabled contexts silently drop metrics.
type noopExporter struct{}

var noop Exporter = &noopExporter{}

func (n *noopExporter) Cmd() string { return valNA }

func (n *noopExporter) Set(_ context.Context, _, _ string, _ float64, _ string)       {}
func (n *noopExporter) Incr(_ context.Context, _, _, _ string)                        {}
func (n *noopExporter) Decr(_ context.Context, _, _, _ string)                        {}
func (n *noopExporter) Count(_ context.Context, _, _, _ string)                       {}
func (n *noopExporter) CountDelta(_ context.Context, _, _ string, _ uint64, _ string) {}
func (n *noopExporter) Observe(_ context.Context, _, _ string, _ float64)             {}
func (n *noopExporter) Sample(_ context.Context, _, _ string, _ float64, _ string)    {}
