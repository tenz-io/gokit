package monitor

import "github.com/prometheus/client_golang/prometheus"

// Label names shared by every metric family. V3 unifies the label set across
// counter/gauge/histogram/summary (v2's histogram dropped `opt`, which made
// cardinality strategy inconsistent); every family now carries the same four
// dimensions so dashboards and queries are uniform.
const (
	labelCmd   = "cmd"
	labelDsCmd = "dsCmd"
	labelCode  = "code"
	labelOpt   = "opt"

	// defaultNamespace/Subsystem are the metric prefix; overridable via Init.
	defaultNamespace = "gokit"
	defaultSubsystem = "flight"

	// valNA is the placeholder for empty string labels, so Prometheus never
	// sees an unset label value (which would silently split series).
	valNA = "NA"

	// codeOK / codeErr are the only two result codes retained after
	// normalization, keeping cardinality bounded regardless of how callers
	// spell their error codes.
	codeOK  = "0"
	codeErr = "1"

	// optActive is the active-request gauge's opt slot: the gauge that tracks
	// in-flight calls uses opt="actives" so it does not collide with business
	// opt dimensions like hit/miss.
	optActive = "actives"
)

// latencyBuckets is the histogram bucket layout for latency in milliseconds,
// spanning 0.1ms to 10s. Reused from v2 (battle-tested across the codebase).
var latencyBuckets = []float64{
	1e-1,     // 0.1ms  factor 10
	1e0, 3e0, // 1ms    factor 3
	1e1, 2e1, 4e1, 8e1, // 10ms   factor 2
	1.6e2, 3.2e2, 6.4e2, // 160ms  factor 2
	1e3, 3e3, // 1000ms factor 3
	1e4, // 10000ms to infinite
}

// summaryObjectives configures the quantile targets for the data-size summary.
var summaryObjectives = map[float64]float64{
	0.5:  0.05,
	0.9:  0.01,
	0.95: 0.05,
	0.99: 0.001,
}

// normalizeOpt maps an empty opt to the NA placeholder so that omitted opt
// values collapse into one series instead of many.
func normalizeOpt(opt string) string {
	if opt == "" {
		return valNA
	}
	return opt
}

// normalizeCode collapses any non-zero code to "1" (err) and an empty code to
// "0" (ok). This bounds the cardinality of code across observe/sample: the
// business only ever sees ok/err on latency and size metrics, while exact
// codes are preserved on counter/gauge where they are first-class.
func normalizeCode(code string) string {
	if code == "" || code == codeOK {
		return codeOK
	}
	return codeErr
}

// labelsOf builds the canonical four-dimensional label set for a metric. The
// cmd is the Exporter's scope; dsCmd/code/opt are per-call.
//
// The production hot path uses Exporter.WithLabelValues (a slice fast path with
// no map allocation); labelsOf is kept for tests and tooling that want a
// ready-made prometheus.Labels map for assertions and gathering. It does not
// normalize — pass already-normalized values.
func labelsOf(cmd, dsCmd, code, opt string) prometheus.Labels {
	return prometheus.Labels{
		labelCmd:   cmd,
		labelDsCmd: dsCmd,
		labelCode:  code,
		labelOpt:   opt,
	}
}
