// Package monitor collects Prometheus metrics under a single-flight injection
// model: one request creates (or reuses) a cmd-scoped Exporter and injects it
// into the context; downstream call sites Begin/End recorders against that
// same Exporter. This avoids per-call Init churn and keeps every point on the
// call chain sharing one metric exporter.
//
// Unlike v2, metric families are registered lazily through an explicit
// Configure entry point (with an injectable Registry) rather than a global
// init() that would panic on duplicate registration. Configure is optional:
// callers that never call it get the default Prometheus Registerer as a
// fallback. A Configure called after metrics are already in use fails
// explicitly (returns an error) rather than silently routing points to the
// wrong Registry.
package monitor

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// config holds the resolved registration settings. It is written exactly once
// by Configure and read by every Exporter thereafter.
type config struct {
	namespace string
	subsystem string
	registry  prometheus.Registerer
}

var (
	configMu sync.RWMutex
	cfg      = config{
		namespace: defaultNamespace,
		subsystem: defaultSubsystem,
		registry:  prometheus.DefaultRegisterer,
	}

	// configured is set true by the first successful Configure. It guards
	// against a second Configure silently overwriting an already-applied
	// configuration.
	configured bool

	// metricsMu guards registration of the four metric families. Once
	// registered (against whatever Registerer was active at first use),
	// registerOnce returns the cached instances.
	metricsMu sync.Mutex
	metrics   struct {
		counter   *prometheus.CounterVec
		gauge     *prometheus.GaugeVec
		histogram *prometheus.HistogramVec
		summary   *prometheus.SummaryVec
	}
)

// Option configures the package registration.
type Option func(*config)

// WithNamespace overrides the metric namespace (default "gokit").
func WithNamespace(ns string) Option {
	return func(c *config) { c.namespace = ns }
}

// WithSubsystem overrides the metric subsystem (default "flight").
func WithSubsystem(sub string) Option {
	return func(c *config) { c.subsystem = sub }
}

// WithRegistry injects a custom Prometheus Registerer. Use a private
// registry in tests to avoid colliding with the global default; use a
// prometheus.NewRegistry() in production to keep these metrics out of the
// default /metrics scrape if desired.
func WithRegistry(r prometheus.Registerer) Option {
	return func(c *config) { c.registry = r }
}

// ErrAlreadyConfigured is returned by Configure when called more than once.
// The first successful Configure pins the registration settings for the life
// of the process; a second call cannot change where points are routed.
var ErrAlreadyConfigured = errors.New("monitor: Configure already called; registration is fixed")

// ErrAlreadyInUse is returned by Configure when called after an Exporter has
// already been built (and thus the metric families have been registered
// against whatever Registry was active — by default the Prometheus default).
// Call Configure at process start, before any NewExporter/Init/Begin.
var ErrAlreadyInUse = errors.New("monitor: metrics already registered; call Configure before first use")

// Configure sets the package's metric registration (Registry, namespace,
// subsystem). Call it once at process start, before any NewExporter/Init/Begin.
//
// Configure is optional: if never called, metrics register against
// prometheus.DefaultRegisterer on first Exporter construction.
//
// Configure fails explicitly — returns a non-nil error — rather than silently
// doing nothing when it is too late:
//
//   - ErrAlreadyInUse: an Exporter was already built, so the metric families
//     are live against the active (default or prior-Configure) Registry.
//     Routing points to a new Registry now would split metrics silently.
//   - ErrAlreadyConfigured: Configure was called a second time but no
//     Exporter was built in between. The first call's settings are kept.
//
// Returning the error lets the caller decide (log loud, fail fast) instead of
// discovering stale metrics in a dashboard later.
func Configure(opts ...Option) error {
	configMu.Lock()
	defer configMu.Unlock()

	// If metrics are already live, the active cfg (default or a prior
	// Configure's) has been committed to a Registry. Refuse to change it.
	// Check this before the configured flag so an Exporter built between a
	// first Configure and a second one surfaces as ErrAlreadyInUse (the more
	// precise diagnosis), not ErrAlreadyConfigured.
	metricsMu.Lock()
	alreadyInUse := metrics.counter != nil
	metricsMu.Unlock()
	if alreadyInUse {
		return ErrAlreadyInUse
	}
	if configured {
		return ErrAlreadyConfigured
	}

	next := cfg
	for _, opt := range opts {
		if opt != nil {
			opt(&next)
		}
	}
	if next.namespace == "" {
		next.namespace = defaultNamespace
	}
	if next.subsystem == "" {
		next.subsystem = defaultSubsystem
	}
	if next.registry == nil {
		next.registry = prometheus.DefaultRegisterer
	}
	cfg = next
	configured = true
	return nil
}

// registerOnce registers all four metric families against the active Registerer
// on first call and returns the cached collectors. The families share the same
// four-label set {cmd, dsCmd, code, opt} (see labels.go) so label cardinality
// is uniform across counter/gauge/histogram/summary.
//
// registerOnce is safe under concurrent first-use from multiple goroutines.
func registerOnce() {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	if metrics.counter != nil {
		return
	}

	configMu.RLock()
	c := cfg
	configMu.RUnlock()

	metrics.counter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: c.namespace,
		Subsystem: c.subsystem,
		Name:      "singleFlightC",
		Help:      "single flight counter tracking",
	}, []string{labelCmd, labelDsCmd, labelCode, labelOpt})

	metrics.gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: c.namespace,
		Subsystem: c.subsystem,
		Name:      "singleFlightG",
		Help:      "single flight gauge tracking",
	}, []string{labelCmd, labelDsCmd, labelCode, labelOpt})

	metrics.histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: c.namespace,
		Subsystem: c.subsystem,
		Name:      "singleFlightH",
		Buckets:   latencyBuckets,
		Help:      "single flight histogram tracking",
	}, []string{labelCmd, labelDsCmd, labelCode, labelOpt})

	metrics.summary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  c.namespace,
		Subsystem:  c.subsystem,
		Name:       "singleFlightS",
		Objectives: summaryObjectives,
		Help:       "single flight summary tracking",
	}, []string{labelCmd, labelDsCmd, labelCode, labelOpt})

	if c.registry != nil {
		c.registry.MustRegister(
			metrics.counter,
			metrics.gauge,
			metrics.histogram,
			metrics.summary,
		)
	}
}

// snapshotMetrics returns the registered metric families, registering them on
// first use. Exporter calls this once per process at construction time.
func snapshotMetrics() (counter *prometheus.CounterVec, gauge *prometheus.GaugeVec,
	histogram *prometheus.HistogramVec, summary *prometheus.SummaryVec) {
	registerOnce()
	metricsMu.Lock()
	defer metricsMu.Unlock()
	return metrics.counter, metrics.gauge, metrics.histogram, metrics.summary
}

// resetForTest clears the package registration state so a test can call
// Configure with a fresh Registry and exercise registration in isolation.
// It is only safe to use from tests; production code must not call it once any
// Exporter has been published to a context that other goroutines may read.
func resetForTest() {
	configMu.Lock()
	cfg = config{
		namespace: defaultNamespace,
		subsystem: defaultSubsystem,
		registry:  prometheus.DefaultRegisterer,
	}
	configured = false
	configMu.Unlock()

	metricsMu.Lock()
	metrics = struct {
		counter   *prometheus.CounterVec
		gauge     *prometheus.GaugeVec
		histogram *prometheus.HistogramVec
		summary   *prometheus.SummaryVec
	}{}
	metricsMu.Unlock()
}
