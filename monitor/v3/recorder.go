package monitor

import (
	"context"
	"time"
)

// Recorder times a single call: Begin records the start time and increments
// the active-request gauge; End records the elapsed time into the latency
// histogram, increments the counter, and decrements the active-request gauge.
//
// End is synchronous (v2 fired a goroutine per End, which lost ordering and
// could unbalance the active gauge). To stay balanced End is also
// idempotent: a second End is a no-op, so a deferred End plus an explicit one
// never double-decrements the active gauge.
//
// Recorder captures the context from Begin in a cancellation-detached form
// (context.WithoutCancel) and passes it to every End-side Exporter call, so a
// custom Exporter can still read trace IDs, tenant IDs, or attach exemplars
// even after the request context has been cancelled — without End blocking
// on a cancelled context.
type Recorder struct {
	exp     Exporter
	dsCmd   string
	start   time.Time
	ctx     context.Context
	endOnce endGuard
}

// endGuard makes End idempotent without importing sync/Once (whose Do is
// heavier than needed here and would make the hot path non-inlineable).
// Reads/writes are protected by the contained done bool; Recorder is only
// used within a single goroutine's Begin/End scope, so the guard does not
// need to be atomic across goroutines — see the End comment.
type endGuard struct{ done bool }

// Begin returns a Recorder for dsCmd that uses the Exporter injected into ctx
// (or a no-op if none). Begin increments the active-request gauge immediately;
// pair it with exactly one End (defer rec.EndWithError(err) is the usual form).
//
// Begin is the single-flight entry point for downstream call sites: every
// Recorder along the chain shares the one Exporter Init put into the context.
func Begin(ctx context.Context, dsCmd string) *Recorder {
	exp := FromContext(ctx)
	exp.Incr(ctx, dsCmd, codeOK, optActive)
	return &Recorder{
		exp:   exp,
		dsCmd: dsCmd,
		start: time.Now(),
		// Detach cancellation but keep all values (trace, tenant, ...) so End
		// can still enrich metrics after the request context is done.
		ctx: withoutCancel(ctx),
	}
}

// End ends the record with the default ok code. Idempotent.
func (r *Recorder) End() {
	r.EndWithCodeOpt(codeOK, valNA)
}

// EndWithCode ends the record with the given code. Idempotent.
func (r *Recorder) EndWithCode(code string) {
	r.EndWithCodeOpt(code, valNA)
}

// EndWithOpt ends the record with the given opt and the default ok code.
// Idempotent.
func (r *Recorder) EndWithOpt(opt string) {
	r.EndWithCodeOpt(codeOK, opt)
}

// EndWithError ends the record mapping err to a code: nil → ok, non-nil → err.
// Idempotent.
func (r *Recorder) EndWithError(err error) {
	r.EndWithErrorOpt(err, valNA)
}

// EndWithErrorOpt is EndWithError with an extra opt label. Idempotent.
func (r *Recorder) EndWithErrorOpt(err error, opt string) {
	code := codeOK
	if err != nil {
		code = codeErr
	}
	r.EndWithCodeOpt(code, opt)
}

// EndWithCodeOpt is the terminal End: it records the elapsed latency into the
// histogram, increments the counter, and decrements the active gauge — all
// synchronously and exactly once. A second call is a no-op.
//
// Recorder is scoped to one goroutine's Begin/End, so the idempotency guard
// does not need cross-goroutine visibility; callers must not share a Recorder
// across goroutines (Begin a new one per goroutine instead).
func (r *Recorder) EndWithCodeOpt(code, opt string) {
	if r == nil || r.endOnce.done {
		return
	}
	r.endOnce.done = true

	// Normalize the opt once; Observe normalizes the code itself.
	nopt := normalizeOpt(opt)

	durMillis := asMillis(r.start)

	// r.ctx is the cancellation-detached view of Begin's context; it carries
	// trace/tenant values for custom Exporters without blocking on cancel.
	r.exp.Observe(r.ctx, r.dsCmd, code, durMillis)
	r.exp.Count(r.ctx, r.dsCmd, code, nopt)
	r.exp.Decr(r.ctx, r.dsCmd, codeOK, optActive)
}

// asMillis returns the milliseconds elapsed since begin.
func asMillis(begin time.Time) float64 {
	return float64(time.Since(begin).Nanoseconds()) / 1e6
}

// withoutCancel returns a context carrying ctx's values but no deadline or
// cancellation channel. It is a nil-safe wrapper around context.WithoutCancel
// (Go 1.21+): passing a nil ctx returns a fresh background context.
func withoutCancel(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}
