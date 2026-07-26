package monitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// freshRegistry returns a new Registry for an isolated test, and resets the
// package registration so the test's Configure call registers against it
// instead of the process default. A *prometheus.Registry is both a
// Registerer and a Gatherer, so tests pass it straight through.
func freshRegistry(t *testing.T) *prometheus.Registry {
	t.Helper()
	reg := prometheus.NewRegistry()
	resetForTest()
	if err := Configure(WithRegistry(reg)); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	return reg
}

func TestConfigureFirstCallWinsSecondFails(t *testing.T) {
	regA := prometheus.NewRegistry()
	regB := prometheus.NewRegistry()
	resetForTest()

	// First Configure wins.
	if err := Configure(WithRegistry(regA), WithNamespace("svcA"), WithSubsystem("flA")); err != nil {
		t.Fatalf("first Configure: %v", err)
	}
	// Second Configure must fail explicitly: namespace/registry stay as A.
	err := Configure(WithRegistry(regB), WithNamespace("svcB"), WithSubsystem("flB"))
	if !errors.Is(err, ErrAlreadyConfigured) {
		t.Fatalf("second Configure err = %v, want ErrAlreadyConfigured", err)
	}

	configMu.RLock()
	defer configMu.RUnlock()
	if cfg.namespace != "svcA" {
		t.Errorf("namespace = %q, want svcA (first Configure wins)", cfg.namespace)
	}
	if cfg.subsystem != "flA" {
		t.Errorf("subsystem = %q, want flA", cfg.subsystem)
	}
	if cfg.registry != regA {
		t.Errorf("registry changed after failed second Configure (should be sticky)")
	}
}

func TestConfigureAfterFirstUseFails(t *testing.T) {
	// Start from a configured private Registry (so we never touch the
	// process-default Registry, which other tests may have populated).
	regA := prometheus.NewRegistry()
	resetForTest()
	if err := Configure(WithRegistry(regA)); err != nil {
		t.Fatalf("first Configure: %v", err)
	}

	// Build an Exporter: the metric families go live against regA.
	NewExporter("svc")

	// Now a late Configure to a *different* Registry must fail loudly,
	// rather than silently routing subsequent points to the wrong Registry.
	regB := prometheus.NewRegistry()
	err := Configure(WithRegistry(regB))
	if !errors.Is(err, ErrAlreadyInUse) {
		t.Fatalf("late Configure err = %v, want ErrAlreadyInUse", err)
	}

	// And the active Registry must still be regA (unchanged).
	configMu.RLock()
	gotReg := cfg.registry
	configMu.RUnlock()
	if gotReg != regA {
		t.Error("late Configure changed the active Registry despite failing")
	}
}

func TestNewExporterEmptyCmdNormalizes(t *testing.T) {
	freshRegistry(t)
	exp := NewExporter("")
	if got := exp.Cmd(); got != valNA {
		t.Errorf("NewExporter(\"\").Cmd() = %q, want %q", got, valNA)
	}
	exp2 := NewExporter("svc")
	if got := exp2.Cmd(); got != "svc" {
		t.Errorf("NewExporter(\"svc\").Cmd() = %q, want svc", got)
	}
}

func TestBeginEndActiveGaugeBalances(t *testing.T) {
	reg := freshRegistry(t)
	exp := NewExporter("svc")

	// Run several Begin/End pairs and assert the active gauge returns to 0
	// (every Inc matched a Dec), and the counter equals the number of Ends.
	for i := 0; i < 5; i++ {
		rec := newRecorderForTest(exp, "op")
		rec.End()
	}

	gotActive := gaugeValue(t, reg, map[string]string{
		labelCmd: "svc", labelDsCmd: "op", labelCode: codeOK, labelOpt: optActive,
	})
	if gotActive != 0 {
		t.Errorf("active gauge = %v after balanced Begin/End, want 0", gotActive)
	}
	gotCount := counterValue(t, reg, map[string]string{
		labelCmd: "svc", labelDsCmd: "op", labelCode: codeOK, labelOpt: valNA,
	})
	if gotCount != 5 {
		t.Errorf("counter = %v after 5 Ends, want 5", gotCount)
	}
}

func TestEndIdempotentDoesNotDoubleDecrement(t *testing.T) {
	reg := freshRegistry(t)
	exp := NewExporter("svc")

	rec := newRecorderForTest(exp, "op")
	rec.End()
	rec.End() // second End must be a no-op

	gotActive := gaugeValue(t, reg, map[string]string{
		labelCmd: "svc", labelDsCmd: "op", labelCode: codeOK, labelOpt: optActive,
	})
	if gotActive != 0 {
		t.Errorf("active gauge = %v after double End, want 0 (idempotent)", gotActive)
	}
	gotCount := counterValue(t, reg, map[string]string{
		labelCmd: "svc", labelDsCmd: "op", labelCode: codeOK, labelOpt: valNA,
	})
	if gotCount != 1 {
		t.Errorf("counter = %v after double End, want 1 (idempotent)", gotCount)
	}
}

func TestEndWithErrorMapsCode(t *testing.T) {
	reg := freshRegistry(t)
	exp := NewExporter("svc")

	// success → code 0
	newRecorderForTest(exp, "ok").EndWithError(nil)
	// failure → code 1
	newRecorderForTest(exp, "err").EndWithError(errors.New("boom"))

	countOk := counterValue(t, reg, map[string]string{
		labelCmd: "svc", labelDsCmd: "ok", labelCode: codeOK, labelOpt: valNA,
	})
	if countOk != 1 {
		t.Errorf("ok counter = %v, want 1", countOk)
	}
	countErr := counterValue(t, reg, map[string]string{
		labelCmd: "svc", labelDsCmd: "err", labelCode: codeErr, labelOpt: valNA,
	})
	if countErr != 1 {
		t.Errorf("err counter = %v, want 1", countErr)
	}
}

func TestObserveNormalizesCodeForBoundedCardinality(t *testing.T) {
	reg := freshRegistry(t)
	exp := NewExporter("svc")

	// Three different non-zero codes must collapse onto the single err series.
	exp.Observe(context.Background(), "op", "2", 1.0)
	exp.Observe(context.Background(), "op", "500", 2.0)
	exp.Observe(context.Background(), "op", "panic", 3.0)

	count := histogramSampleCount(t, reg, map[string]string{
		labelCmd: "svc", labelDsCmd: "op", labelCode: codeErr, labelOpt: valNA,
	})
	if count != 3 {
		t.Errorf("err histogram sample count = %d, want 3 (non-zero codes collapsed)", count)
	}

	// The ok series should not exist at all (no zero-code observe happened).
	if _, ok := findMetric(t, reg, "gokit_flight_singleFlightH", map[string]string{
		labelCmd: "svc", labelDsCmd: "op", labelCode: codeOK, labelOpt: valNA,
	}); ok {
		t.Error("ok histogram series exists; expected only the err series")
	}
}

// ---- helpers ----

// findMetric returns the gathered dto.Metric matching name+labels from reg,
// plus whether such a series exists.
func findMetric(t *testing.T, reg *prometheus.Registry, name string,
	labels map[string]string) (*dto.Metric, bool) {
	t.Helper()
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.GetMetric() {
			if matchLabels(m.GetLabel(), labels) {
				return m, true
			}
		}
	}
	return nil, false
}

func counterValue(t *testing.T, reg *prometheus.Registry, labels map[string]string) float64 {
	t.Helper()
	m, ok := findMetric(t, reg, "gokit_flight_singleFlightC", labels)
	if !ok {
		t.Fatalf("counter %v not found", labels)
	}
	return m.GetCounter().GetValue()
}

func gaugeValue(t *testing.T, reg *prometheus.Registry, labels map[string]string) float64 {
	t.Helper()
	m, ok := findMetric(t, reg, "gokit_flight_singleFlightG", labels)
	if !ok {
		t.Fatalf("gauge %v not found", labels)
	}
	return m.GetGauge().GetValue()
}

func histogramSampleCount(t *testing.T, reg *prometheus.Registry, labels map[string]string) uint64 {
	t.Helper()
	m, ok := findMetric(t, reg, "gokit_flight_singleFlightH", labels)
	if !ok {
		t.Fatalf("histogram %v not found", labels)
	}
	return m.GetHistogram().GetSampleCount()
}

func matchLabels(got []*dto.LabelPair, want map[string]string) bool {
	if len(got) != len(want) {
		return false
	}
	for _, p := range got {
		if want[p.GetName()] != p.GetValue() {
			return false
		}
	}
	return true
}

// newRecorderForTest builds a Recorder against a specific Exporter without
// touching a context, for unit tests that exercise Begin/End mechanics.
func newRecorderForTest(exp Exporter, dsCmd string) *Recorder {
	exp.Incr(context.Background(), dsCmd, codeOK, optActive)
	return &Recorder{exp: exp, dsCmd: dsCmd, start: time.Now(), ctx: context.Background()}
}
