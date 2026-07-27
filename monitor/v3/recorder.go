package monitor

import (
	"context"
	"time"
)

// Recorder 为单次调用计时:Begin 记录起始时间并递增活跃请求 gauge;
// End 将耗时记录到延迟 histogram 中,递增 counter,并递减活跃请求
// gauge。
//
// End 是同步的(v2 为每个 End 起一个 goroutine,丢失了顺序并可能使
// 活跃 gauge 失衡)。为保持平衡,End 也是幂等的:第二次 End 是
// no-op,因此 defer 的 End 与显式的 End 组合也不会重复递减活跃
// gauge。
//
// Recorder 在 Begin 时以一种与取消解耦的形态(context.WithoutCancel)
// 捕获 context,并将其传给 End 侧的每次 Exporter 调用,这样自定义
// Exporter 即便在请求 context 已被取消之后仍可读取 trace ID、tenant
// ID 或附加 exemplar —— 而 End 不会阻塞在被取消的 context 上。
type Recorder struct {
	exp     Exporter
	dsCmd   string
	start   time.Time
	ctx     context.Context
	endOnce endGuard
}

// endGuard 使 End 幂等,而无需引入 sync/Once(其 Do 比此处所需更重,
// 且会让热路径无法内联)。读写由所包含的 done bool 保护;Recorder
// 仅在单个 goroutine 的 Begin/End 作用域内使用,因此该守卫不需要
// 跨 goroutine 的原子性 —— 见 End 的注释。
type endGuard struct{ done bool }

// Begin 返回一个用于 dsCmd 的 Recorder,它使用注入到 ctx 中的
// Exporter(若没有则为 no-op)。Begin 立即递增活跃请求 gauge;请与
// 恰好一次 End 配对(defer rec.EndWithError(err) 是常见写法)。
//
// Begin 是下游调用点的 single-flight 入口:链路上每个 Recorder 都共享
// Init 放入 context 的那一个 Exporter。
func Begin(ctx context.Context, dsCmd string) *Recorder {
	exp := FromContext(ctx)
	exp.Incr(ctx, dsCmd, codeOK, optActive)
	return &Recorder{
		exp:   exp,
		dsCmd: dsCmd,
		start: time.Now(),
		// Detach cancellation 但保留全部值(trace、tenant……),这样
		// End 在请求 context 结束后仍能为 metric 附加信息。
		ctx: withoutCancel(ctx),
	}
}

// End 以默认的 ok code 结束记录。幂等。
func (r *Recorder) End() {
	r.EndWithCodeOpt(codeOK, valNA)
}

// EndWithCode 以给定 code 结束记录。幂等。
func (r *Recorder) EndWithCode(code string) {
	r.EndWithCodeOpt(code, valNA)
}

// EndWithOpt 以给定 opt 与默认 ok code 结束记录。幂等。
func (r *Recorder) EndWithOpt(opt string) {
	r.EndWithCodeOpt(codeOK, opt)
}

// EndWithError 结束记录并将 err 映射为 code:nil → ok,非 nil → err。
// 幂等。
func (r *Recorder) EndWithError(err error) {
	r.EndWithErrorOpt(err, valNA)
}

// EndWithErrorOpt 是 EndWithError 加一个额外 opt label 的版本。幂等。
func (r *Recorder) EndWithErrorOpt(err error, opt string) {
	code := codeOK
	if err != nil {
		code = codeErr
	}
	r.EndWithCodeOpt(code, opt)
}

// EndWithCodeOpt 是终态的 End:它将耗时记录到 histogram,递增
// counter,并递减活跃 gauge —— 全部同步执行且仅一次。第二次调用为
// no-op。
//
// Recorder 的作用域局限于单个 goroutine 的 Begin/End,因此幂等守卫
// 不需要跨 goroutine 可见性;调用方不得跨 goroutine 共享同一个
// Recorder(应在每个 goroutine 中各自 Begin 一个新的)。
func (r *Recorder) EndWithCodeOpt(code, opt string) {
	if r == nil || r.endOnce.done {
		return
	}
	r.endOnce.done = true

	// 归一化 opt 仅一次;Observe 自行归一化 code。
	nopt := normalizeOpt(opt)

	durMillis := asMillis(r.start)

	// r.ctx 是 Begin context 的与取消解耦的视图;它携带 trace/tenant
	// 值供自定义 Exporter 使用,而不会阻塞在 cancel 上。
	r.exp.Observe(r.ctx, r.dsCmd, code, durMillis)
	r.exp.Count(r.ctx, r.dsCmd, code, nopt)
	r.exp.Decr(r.ctx, r.dsCmd, codeOK, optActive)
}

// asMillis 返回自 begin 以来经过的毫秒数。
func asMillis(begin time.Time) float64 {
	return float64(time.Since(begin).Nanoseconds()) / 1e6
}

// withoutCancel 返回一个携带 ctx 的值但没有 deadline 或取消 channel 的
// context。它是 context.WithoutCancel(Go 1.21+)的 nil 安全封装:
// 传入 nil ctx 时返回一个全新的 background context。
func withoutCancel(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}
