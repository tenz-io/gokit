package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type singleFlightCtxKeyType string

const (
	defaultNamespace = "trackingo"
	defaultSubsystem = "flight"
	activeKey        = "actives"
)

const (
	defaultMetricVal = "NA"
	defaultCodeErr   = "1"
	defaultCodeOk    = "0"
)

const (
	singleFlightCtxKey = singleFlightCtxKeyType("singleFlight_ctx_key")
)

var (
	latencyBuckets = []float64{
		1e-1,     //0.1ms factor 10
		1e0, 3e0, //1ms factor 3
		1e1, 2e1, 4e1, 8e1, //10ms factor 2
		1.6e2, 3.2e2, 6.4e2, //160ms factor 2
		1e3, 3e3, //1000ms factor 3
		1e4, //10000ms to infinite
	}
	summaryObjectives = map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.95: 0.05,
		0.99: 0.001,
	}
)

var (
	singleFlightCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: defaultNamespace,
		Subsystem: defaultSubsystem,
		Name:      "singleFlightC",
		Help:      "single flight counter tracking",
	}, []string{"cmd", "dsCmd", "code", "opt"})

	singleFlightGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: defaultNamespace,
		Subsystem: defaultSubsystem,
		Name:      "singleFlightG",
		Help:      "single flight gauge tracking",
	}, []string{"cmd", "dsCmd", "code", "opt"})

	singleFlightHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: defaultNamespace,
		Subsystem: defaultSubsystem,
		Name:      "singleFlightH",
		Buckets:   latencyBuckets,
		Help:      "single flight histogram tracking",
	}, []string{"cmd", "dsCmd", "code"})

	singleFlightSummary = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  defaultNamespace,
		Subsystem:  defaultSubsystem,
		Objectives: summaryObjectives,
		Name:       "singleFlightS",
		Help:       "single flight summary tracking",
	}, []string{"cmd", "dsCmd", "code", "opt"})
)

func init() {
	prometheus.MustRegister(
		singleFlightGauge,
		singleFlightHistogram,
		singleFlightCounter,
		singleFlightSummary,
	)
}

// SingleFlight is the interface for single flight monitor
//
//go:generate mockery --name SingleFlight --filename singleflight_mock.go --inpackage
type SingleFlight interface {
	// Set in gauge
	Set(ctx context.Context, dsCmd string, code string, val float64, opt string)
	// Incr in gauge
	Incr(ctx context.Context, dsCmd string, code string, opt string)
	// Decr in gauge
	Decr(ctx context.Context, dsCmd string, code string, opt string)
	// Count in counter with delta 1
	Count(ctx context.Context, dsCmd string, code string, opt string)
	// CountDelta in counter with delta
	CountDelta(ctx context.Context, dsCmd string, code string, delta int, opt string)
	// Observe in histogram, usually for latency
	Observe(ctx context.Context, dsCmd string, code string, millis float64)
	// Sample in summary, usually for data size
	Sample(ctx context.Context, dsCmd string, code string, val float64, opt string)
	// BeginRecord start a recorder
	BeginRecord(ctx context.Context, dsCmd string) *Recorder
}

// Recorder is the recorder for single flight monitor
// Use BeginRecord to create a recorder, it will record the start time
// Use End to end the recorder, it will calculate the duration and record the metrics
type Recorder struct {
	singleFlight SingleFlight
	ctx          context.Context
	dsCmd        string
	startTime    time.Time
}

func newRecorder(singleFlight SingleFlight, ctx context.Context, dsCmd string) *Recorder {
	singleFlight.Incr(ctx, dsCmd, defaultCodeOk, activeKey)
	return &Recorder{
		singleFlight: singleFlight,
		ctx:          ctx,
		dsCmd:        dsCmd,
		startTime:    time.Now(),
	}
}

// End the recorder with default code 0
func (r *Recorder) End() {
	r.EndWithCode(defaultCodeOk)
}

// EndWithCode end the recorder with code
func (r *Recorder) EndWithCode(code string) {
	r.EndWithCodeOpt(code, "")
}

// EndWithOpt end the recorder with opt and default code 0
func (r *Recorder) EndWithOpt(opt string) {
	r.EndWithCodeOpt(defaultCodeOk, opt)
}

// EndWithError end the recorder with error
func (r *Recorder) EndWithError(err error) {
	r.EndWithErrorOpt(err, "")
}

// EndWithErrorOpt end the recorder with error and opt
// if error is ValError, use ValError.Code as code
func (r *Recorder) EndWithErrorOpt(err error, opt string) {
	var code = defaultCodeOk

	if err != nil {
		code = defaultCodeErr
	}

	r.EndWithCodeOpt(code, opt)
}

// EndWithCodeOpt end the recorder with code and opt
func (r *Recorder) EndWithCodeOpt(code string, opt string) {
	duringMillis := asMillis(r.startTime)
	go func() {
		r.singleFlight.Count(r.ctx, r.dsCmd, code, opt)
		r.singleFlight.Observe(r.ctx, r.dsCmd, code, duringMillis)
		r.singleFlight.Decr(r.ctx, r.dsCmd, defaultCodeOk, activeKey)
	}()
}

// exporter is the default implementation of SingleFlight
type exporter struct {
	cmd string
}

func NewSingleFlight(cmd string) SingleFlight {
	if cmd == "" {
		cmd = defaultMetricVal
	}

	return &exporter{
		cmd: cmd,
	}
}

// getSimplePromLabels get simple prometheus labels
// labels: cmd, dsCmd, code
func (e *exporter) getSimplePromLabels(dsCmd string, code string) prometheus.Labels {
	labels := prometheus.Labels{
		"cmd":   e.cmd,
		"dsCmd": dsCmd,
		"code":  code,
	}

	return labels
}

// getFullPromLabels get full prometheus labels
// labels: cmd, dsCmd, code, opt
func (e *exporter) getFullPromLabels(dsCmd string, code string, opt string) prometheus.Labels {
	labels := e.getSimplePromLabels(dsCmd, code)
	labels["opt"] = opt

	return labels
}

func (e *exporter) Set(ctx context.Context, dsCmd string, code string, val float64, opt string) {
	if opt == "" {
		opt = defaultMetricVal
	}

	labels := e.getFullPromLabels(dsCmd, code, opt)

	singleFlightGauge.With(labels).Set(val)
}

func (e *exporter) Incr(ctx context.Context, dsCmd string, code string, opt string) {
	if opt == "" {
		opt = defaultMetricVal
	}

	labels := e.getFullPromLabels(dsCmd, code, opt)
	singleFlightGauge.With(labels).Inc()
}

func (e *exporter) Decr(ctx context.Context, dsCmd string, code string, opt string) {
	if opt == "" {
		opt = defaultMetricVal
	}

	labels := e.getFullPromLabels(dsCmd, code, opt)
	singleFlightGauge.With(labels).Dec()
}

func (e *exporter) Count(ctx context.Context, dsCmd string, code string, opt string) {
	if opt == "" {
		opt = defaultMetricVal
	}

	labels := e.getFullPromLabels(dsCmd, code, opt)
	singleFlightCounter.With(labels).Inc()
}

func (e *exporter) CountDelta(ctx context.Context, dsCmd string, code string, delta int, opt string) {
	if opt == "" {
		opt = defaultMetricVal
	}

	labels := e.getFullPromLabels(dsCmd, code, opt)
	singleFlightCounter.With(labels).Add(float64(delta))
}

func (e *exporter) Sample(ctx context.Context, dsCmd string, code string, val float64, opt string) {
	// reduce prometheus export data amount
	// mapping non-zero code to 1
	if code != defaultCodeOk {
		code = defaultCodeErr
	}

	if opt == "" {
		opt = defaultMetricVal
	}

	labels := e.getFullPromLabels(dsCmd, code, opt)
	singleFlightSummary.With(labels).Observe(val)
}

func (e *exporter) Observe(ctx context.Context, dsCmd string, code string, millis float64) {
	// reduce prometheus export data amount
	// mapping non-zero code to 1
	if code != defaultCodeOk {
		code = defaultCodeErr
	}
	labels := e.getSimplePromLabels(dsCmd, code)
	singleFlightHistogram.With(labels).Observe(millis)
}

func (e *exporter) BeginRecord(ctx context.Context, dsCmd string) *Recorder {
	return newRecorder(e, ctx, dsCmd)
}

type empty struct {
}

func (e *empty) Set(ctx context.Context, dsCmd string, code string, val float64, opt string) {
}

func (e *empty) Incr(ctx context.Context, dsCmd string, code string, opt string) {
}

func (e *empty) Decr(ctx context.Context, dsCmd string, code string, opt string) {
}

func (e *empty) Count(ctx context.Context, dsCmd string, code string, opt string) {
}

func (e *empty) CountDelta(ctx context.Context, dsCmd string, code string, delta int, opt string) {
}

func (e *empty) Sample(ctx context.Context, dsCmd string, code string, val float64, opt string) {
}

func (e *empty) Observe(ctx context.Context, dsCmd string, code string, millis float64) {
}

func (e *empty) BeginRecord(ctx context.Context, dsCmd string) *Recorder {
	return newRecorder(e, ctx, dsCmd)
}

func asMillis(begin time.Time) float64 {
	return float64(time.Now().Sub(begin).Nanoseconds()) / 1e6
}

// FromContext get single flight monitor from ctx
// return empty monitor if not found, always not be nil
func FromContext(ctx context.Context) SingleFlight {
	mon, _ := fromContext(ctx)
	return mon
}

// HasSingleFlight check if ctx has single flight monitor
func HasSingleFlight(ctx context.Context) bool {
	if _, err := fromContext(ctx); err != nil {
		return false
	}
	return true
}

// fromContext get single flight monitor from ctx
// return error and empty monitor if not found
// return empty monitor if found but wrong type
// so this func will always return a monitor
func fromContext(ctx context.Context) (SingleFlight, error) {
	data := ctx.Value(singleFlightCtxKey)
	if data == nil {
		return &empty{}, fmt.Errorf("single flight monitor not found in the ctx")
	}
	mon, ok := data.(SingleFlight)
	if ok == false {
		return &empty{}, fmt.Errorf("wrong value type for single flight monitor key, value:[%v]", data)
	}
	return mon, nil
}

// WithMonitor inject single flight monitor to ctx
func WithMonitor(ctx context.Context, singleFlight SingleFlight) context.Context {
	ctx = context.WithValue(ctx, singleFlightCtxKey, singleFlight)
	return ctx
}

// BeginRecord start a recorder
func BeginRecord(ctx context.Context, dsCmd string) *Recorder {
	return FromContext(ctx).BeginRecord(ctx, dsCmd)
}

// InitSingleFlight init single flight monitor in ctx
// if ctx already has single flight monitor, return ctx directly
func InitSingleFlight(ctx context.Context, cmd string) context.Context {
	var mon SingleFlight
	var err error
	if mon, err = fromContext(ctx); err != nil {
		mon = NewSingleFlight(cmd)
	}
	return WithMonitor(ctx, mon)
}

// CopyToContext copy single flight monitor from src ctx to dst ctx
func CopyToContext(srcCtx, dstCtx context.Context) context.Context {
	if srcCtx == nil || dstCtx == nil {
		return dstCtx
	}

	singleFlight, err := fromContext(srcCtx)
	if err != nil {
		return srcCtx
	}

	dstCtx = WithMonitor(dstCtx, singleFlight)
	return dstCtx
}
