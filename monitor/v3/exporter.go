package monitor

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter 是 cmd 级别的 metric exporter 接口。每个入口命令一个
// Exporter(由 Init 创建或复用并注入到 context 中),沿调用链的所有
// 下游 Begin/End 共享它 —— 这就是 single-flight 注入模型。所有采样
// 方法都是同步且轻分配的:它们归一化 label 值(空 opt → NA,非零
// code → err)并调用 WithLabelValues,即 Prometheus 的 slice 快速
// 路径,因此热路径上不分配 map。
//
// Exporter 被设计为接口,使 FromContext 在未注入任何 Exporter 时可
// 返回一个非 nil 的 no-op 实现,让调用方无需 nil 检查。
type Exporter interface {
	// Cmd 返回所绑定的 cmd label 值。
	Cmd() string

	// Set / Incr / Decr 操作 gauge(瞬时值、活跃计数)。
	Set(ctx context.Context, dsCmd, code string, val float64, opt string)
	Incr(ctx context.Context, dsCmd, code, opt string)
	Decr(ctx context.Context, dsCmd, code, opt string)
	// Count / CountDelta 操作 counter(累计总量)。
	Count(ctx context.Context, dsCmd, code, opt string)
	// CountDelta 向 counter 加上 delta。delta 为 uint64,因为
	// Prometheus counter 是单调非递减的:负 delta 会在
	// prometheus.Counter.Add 内部 panic。无符号类型使该契约在
	// 调用点不可表达(从而不可误用)。
	CountDelta(ctx context.Context, dsCmd, code string, delta uint64, opt string)
	// Observe 将 millis 记录到延迟 histogram;code 被归一化为 ok/err。
	Observe(ctx context.Context, dsCmd, code string, millis float64)
	// Sample 将一个值记录到数据大小 summary;code 被归一化为 ok/err。
	Sample(ctx context.Context, dsCmd, code string, val float64, opt string)
}

// exporter 是基于 Prometheus 的具体 Exporter 实现。它在构造时绑定一个
// cmd 并一次性解析四个 metric family。
type exporter struct {
	cmd       string
	counter   *prometheus.CounterVec
	gauge     *prometheus.GaugeVec
	histogram *prometheus.HistogramVec
	summary   *prometheus.SummaryVec
}

// NewExporter 构造一个绑定到 cmd 的 Exporter。空 cmd 会被归一化为
// NA,使 cmd label 永远有值。metric family 在构造时解析一次(在
// 首次使用时针对当前活跃的 Registerer 注册)。
//
// NewExporter 不要求 Init 已被调用:若未被调用,包会回退到
// prometheus.DefaultRegisterer。
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

// Set 将 gauge 设为绝对值。
func (e *exporter) Set(ctx context.Context, dsCmd, code string, val float64, opt string) {
	e.gauge.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Set(val)
}

// Incr 将 gauge 加 1。
func (e *exporter) Incr(ctx context.Context, dsCmd, code, opt string) {
	e.gauge.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Inc()
}

// Decr 将 gauge 减 1。
func (e *exporter) Decr(ctx context.Context, dsCmd, code, opt string) {
	e.gauge.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Dec()
}

// Count 将 counter 加 1。
func (e *exporter) Count(ctx context.Context, dsCmd, code, opt string) {
	e.counter.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Inc()
}

// CountDelta 向 counter 加上 delta。
func (e *exporter) CountDelta(ctx context.Context, dsCmd, code string, delta uint64, opt string) {
	e.counter.WithLabelValues(e.cmd, dsCmd, code, normalizeOpt(opt)).Add(float64(delta))
}

// Observe 将 millis 记录到延迟 histogram。code 被归一化为 ok/err,
// 因此无论调用方如何书写结果码,延迟的基数都保持受限。
func (e *exporter) Observe(ctx context.Context, dsCmd, code string, millis float64) {
	e.histogram.WithLabelValues(e.cmd, dsCmd, normalizeCode(code), valNA).Observe(millis)
}

// Sample 将 val 记录到数据大小 summary。出于与 Observe 相同的基数
// 考虑,code 被归一化为 ok/err。
func (e *exporter) Sample(ctx context.Context, dsCmd, code string, val float64, opt string) {
	e.summary.WithLabelValues(e.cmd, dsCmd, normalizeCode(code), normalizeOpt(opt)).Observe(val)
}

// noopExporter 是 FromContext 在 context 中未注入 Exporter 时返回的非
// nil Exporter。调用方使用它无需 nil 检查;每个方法都是 no-op,
// 因此未插桩的 context 会静默丢弃 metric。
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
