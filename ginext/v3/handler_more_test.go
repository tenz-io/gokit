package ginext

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tenz-io/gokit/ginext/v3/errcode"
	"github.com/tenz-io/gokit/tracer/v3"
)

// TestChainWith_Order 验证 With 追加的拦截器位于更内层。
func TestChainWith_Order(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	var order []string
	first := Interceptor(func(ctx context.Context, req any, next RpcHandler) (any, error) {
		order = append(order, "first-in")
		resp, err := next(ctx, req)
		order = append(order, "first-out")
		return resp, err
	})
	appended := Interceptor(func(ctx context.Context, req any, next RpcHandler) (any, error) {
		order = append(order, "appended-in")
		resp, err := next(ctx, req)
		order = append(order, "appended-out")
		return resp, err
	})
	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		order = append(order, "handler")
		return "ok", nil
	})

	chain := NewChain(first).With(appended)
	resp, err := chain.Intercept(context.Background(), nil, h)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
	assert.Equal(t,
		[]string{"first-in", "appended-in", "handler", "appended-out", "first-out"},
		order)
}

// TestGetErrCode 验证 err 到码字符串的映射。
func TestGetErrCode(t *testing.T) {
	assert.Equal(t, "0", getErrCode(nil))
	assert.Equal(t, "1", getErrCode(errors.New("any")))

	// errcode.Error 映射为其 Code 字段。
	e := errcode.New(404, "nf").(*errcode.Error)
	assert.Equal(t, "404", getErrCode(e))
}

// TestTracerInterceptor_EnsureRequestID 验证 TracerInterceptor 为空 ctx 补 request ID。
func TestTracerInterceptor_EnsureRequestID(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	// 直接验证 Intercept 通过 TracerInterceptor 后 next 收到的 ctx 带 request ID。
	var seenReqID string
	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		seenReqID = tracer.RequestIDFromCtxOr(ctx)
		return nil, nil
	})

	chain := NewChain(TracerInterceptor())
	_, _ = chain.Intercept(context.Background(), nil, h)
	assert.NotEmpty(t, seenReqID, "expected EnsureRequestID to inject an id")
}

// TestTrafficInterceptor_NilSafe 验证 traffic 未配置时 StartTraffic 返回 nil,
// Interceptor 仍正常透传。
func TestTrafficInterceptor_NilSafe(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	called := false
	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		called = true
		return ginextData{}, nil
	})

	chain := NewChain(TrafficInterceptor())
	resp, err := chain.Intercept(context.Background(), nil, h)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, ginextData{}, resp)
}

// TestSlowLogInterceptor_FastPath 验证未超阈值时不告警(仅执行 handler)。
func TestSlowLogInterceptor_FastPath(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		return "done", nil
	})

	chain := NewChain(SlowLogInterceptor(0)) // 0 → 默认 5s
	resp, err := chain.Intercept(context.Background(), nil, h)
	assert.NoError(t, err)
	assert.Equal(t, "done", resp)
}

// TestPanicRecover_TranslatesError 验证 panic 被翻译成 500 errcode.Error。
func TestPanicRecover_TranslatesError(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		panic("kaboom")
	})

	chain := NewChain(PanicRecoverInterceptor())
	_, err := chain.Intercept(context.Background(), nil, h)
	assert.Error(t, err)

	e, ok := errcode.FromError(err)
	assert.True(t, ok, "panic should translate to errcode.Error")
	if ok {
		assert.NotZero(t, e.Code)
	}
}

// TestChainContext_CmdInCtx 验证 cmd 被透传进 ctx。
func TestChainContext_CmdInCtx(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	ctx := context.Background()
	ctx = withCmd(ctx, "my-rpc")
	assert.Equal(t, "my-rpc", cmdFromCtx(ctx))
}

// TestCmdFromCtx_NilSafety 验证 nil/无 cmd 时返回空串。
func TestCmdFromCtx_NilSafety(t *testing.T) {
	assert.Equal(t, "", cmdFromCtx(nil))
	assert.Equal(t, "", cmdFromCtx(context.Background()))
}

// --- test helpers ---

type ginextData struct{}
