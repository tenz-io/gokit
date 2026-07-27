package ginext

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tenz-io/gokit/monitor/v3"
)

// fakeExporter 记录 Observe/Count 的 (cmd, code) 调用,用于断言 panic
// 被记为失败而非成功。
type fakeExporter struct {
	mu     sync.Mutex
	cmd    string
	obs    []string // Observe 的 code 值(归一化前)
	counts []string // Count 的 code 值
}

func (f *fakeExporter) Cmd() string                                                          { return f.cmd }
func (f *fakeExporter) Set(ctx context.Context, dsCmd, code string, val float64, opt string) {}
func (f *fakeExporter) Incr(ctx context.Context, dsCmd, code, opt string)                    {}
func (f *fakeExporter) Decr(ctx context.Context, dsCmd, code, opt string)                    {}
func (f *fakeExporter) Count(ctx context.Context, dsCmd, code, opt string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.counts = append(f.counts, code)
}
func (f *fakeExporter) CountDelta(ctx context.Context, dsCmd, code string, delta uint64, opt string) {
}
func (f *fakeExporter) Observe(ctx context.Context, dsCmd, code string, millis float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.obs = append(f.obs, code)
}
func (f *fakeExporter) Sample(ctx context.Context, dsCmd, code string, val float64, opt string) {
}

// TestChainPanicRecordedAsFailure 验证调整链顺序后(PanicRecover 最内层),
// handler panic 时 MetricsInterceptor 的 recorder 把这次调用记为失败
// (code 非 "0"),而非误记为成功。
func TestChainPanicRecordedAsFailure(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	fake := &fakeExporter{cmd: "panic-test"}
	// 注入 fake exporter 与 cmd,让 ChainContext 之后的 ctx 用它。
	// 注意:ChainContext 会调 monitor.Init,但 ctx 已有 exporter 时 Init 幂等复用。
	ctx := context.Background()
	ctx = monitor.WithExporter(ctx, fake)
	ctx = withCmd(ctx, "panic-test")

	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		panic("boom")
	})

	_, err := DefaultChain().Intercept(ctx, nil, h)
	assert.Error(t, err)

	fake.mu.Lock()
	defer fake.mu.Unlock()
	assert.NotEmpty(t, fake.obs, "metrics Observe should be called")
	assert.NotEmpty(t, fake.counts, "metrics Count should be called")
	// panic → err 非 nil → EndWithError 的 code 应为 codeErr("1")。
	// Observe 会把 code 归一化,但 Count 传原始 code。
	for _, c := range fake.counts {
		assert.NotEqual(t, "0", c, "panic must not be recorded as success (code 0)")
	}
}

// TestChainSuccessRecordedAsOk 验证正常 handler 时 metrics 记 code=0(对照)。
func TestChainSuccessRecordedAsOk(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	fake := &fakeExporter{cmd: "ok-test"}
	ctx := context.Background()
	ctx = monitor.WithExporter(ctx, fake)
	ctx = withCmd(ctx, "ok-test")

	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	_, err := DefaultChain().Intercept(ctx, nil, h)
	assert.NoError(t, err)

	fake.mu.Lock()
	defer fake.mu.Unlock()
	assert.NotEmpty(t, fake.counts, "Count should be called")
	// 成功 → EndWithError(nil) → code=codeOK("0")。Count 传 "0"。
	sawOk := false
	for _, c := range fake.counts {
		if c == "0" {
			sawOk = true
		}
	}
	assert.True(t, sawOk, "success should record code 0")
}
