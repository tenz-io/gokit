// Package monitor 在 single-flight 注入模型下采集 Prometheus metric:
// 一个请求创建(或复用)一个 cmd 级别的 Exporter 并注入到 context 中,
// 下游调用点基于同一个 Exporter 进行 Begin/End 记录。这避免了每次调用
// 都触发 Init 的抖动,并使调用链上每一个采样点共享同一个 metric exporter。
//
// 与 v2 不同,metric family 通过显式的 Configure 入口(可注入 Registry)
// 进行惰性注册,而非依赖会在重复注册时 panic 的全局 init()。Configure 是
// 可选的:从未调用它的调用方会以默认的 Prometheus Registerer 作为兜底。
// 在 metric 已被使用之后调用 Configure 会显式失败(返回 error),而不是
// 静默地将采样点路由到错误的 Registry。
package monitor

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// config 保存已解析的注册设置。它由 Configure 写入一次,此后由每个
// Exporter 读取。
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

	// configured 在首次成功的 Configure 时置为 true。它用于防止
	// 第二次 Configure 静默覆盖已应用的配置。
	configured bool

	// metricsMu 保护四个 metric family 的注册。一旦注册(针对首次使用时
	// 处于活跃状态的 Registerer),registerOnce 会返回缓存的实例。
	metricsMu sync.Mutex
	metrics   struct {
		counter   *prometheus.CounterVec
		gauge     *prometheus.GaugeVec
		histogram *prometheus.HistogramVec
		summary   *prometheus.SummaryVec
	}
)

// Option 用于配置包的注册。
type Option func(*config)

// WithNamespace 覆盖 metric 的 namespace(默认 "gokit")。
func WithNamespace(ns string) Option {
	return func(c *config) { c.namespace = ns }
}

// WithSubsystem 覆盖 metric 的 subsystem(默认 "flight")。
func WithSubsystem(sub string) Option {
	return func(c *config) { c.subsystem = sub }
}

// WithRegistry 注入一个自定义的 Prometheus Registerer。在测试中使用
// 私有 registry 可避免与全局默认 registry 冲突;在生产环境中如希望这些
// metric 不出现在默认的 /metrics 采集中,可使用 prometheus.NewRegistry()。
func WithRegistry(r prometheus.Registerer) Option {
	return func(c *config) { c.registry = r }
}

// ErrAlreadyConfigured 由重复调用的 Configure 返回。首次成功的 Configure
// 在进程生命周期内固定注册设置;第二次调用无法改变采样点的路由去向。
var ErrAlreadyConfigured = errors.New("monitor: Configure already called; registration is fixed")

// ErrAlreadyInUse 在 Exporter 已被构建(因此 metric family 已注册到当前
// 活跃的 Registry —— 默认即 Prometheus 默认 Registry)之后调用 Configure
// 时返回。请在进程启动时、任何 NewExporter/Init/Begin 之前调用 Configure。
var ErrAlreadyInUse = errors.New("monitor: metrics already registered; call Configure before first use")

// Configure 设置包的 metric 注册(Registry、namespace、subsystem)。请在
// 进程启动时调用一次,在任何 NewExporter/Init/Begin 之前。
//
// Configure 是可选的:若从未调用,metric 会在首次构造 Exporter 时注册到
// prometheus.DefaultRegisterer。
//
// Configure 在为时已晚时会显式失败 —— 返回非 nil 的 error —— 而不是
// 静默地什么也不做:
//
//   - ErrAlreadyInUse:已有 Exporter 被构建,metric family 已对当前活跃的
//     (默认或上一次 Configure 的)Registry 生效。此时将采样点路由到新
//     Registry 会静默地分裂 metric。
//   - ErrAlreadyConfigured:第二次调用 Configure 但期间没有构建任何
//     Exporter。保留首次调用的设置。
//
// 返回 error 让调用方自行决定(大声记日志还是快速失败),而不是事后在
// dashboard 里才发现陈旧的 metric。
func Configure(opts ...Option) error {
	configMu.Lock()
	defer configMu.Unlock()

	// 如果 metric 已生效,当前活跃的 cfg(默认或上一次 Configure 的)已
	// 提交到某个 Registry。拒绝修改。在 configured 标志之前先检查这一项,
	// 这样在首次 Configure 与第二次 Configure 之间构建了 Exporter 的情形
	// 会以 ErrAlreadyInUse(更精确的诊断)暴露,而非 ErrAlreadyConfigured。
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

// registerOnce 在首次调用时将四个 metric family 注册到当前活跃的
// Registerer,并返回已缓存的 collector。这些 family 共享同一套四 label
// 集合 {cmd, dsCmd, code, opt}(见 labels.go),因此 counter/gauge/
// histogram/summary 的 label 基数保持一致。
//
// registerOnce 在多个 goroutine 并发首次使用时是安全的。
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

// snapshotMetrics 返回已注册的 metric family,并在首次使用时完成注册。
// Exporter 在进程构造期调用它一次。
func snapshotMetrics() (counter *prometheus.CounterVec, gauge *prometheus.GaugeVec,
	histogram *prometheus.HistogramVec, summary *prometheus.SummaryVec) {
	registerOnce()
	metricsMu.Lock()
	defer metricsMu.Unlock()
	return metrics.counter, metrics.gauge, metrics.histogram, metrics.summary
}

// resetForTest 清空包的注册状态,使测试可以用全新的 Registry 调用
// Configure 并独立地演练注册流程。它仅可安全地在测试中使用;一旦任何
// Exporter 已发布到可能被其他 goroutine 读取的 context 中,生产代码
// 便不得调用它。
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
