package ginext

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/tenz-io/gokit/logger/v3"
)

func setupLogger(t *testing.T) func(t *testing.T) {
	t.Helper()
	logger.ConfigureWithOpts(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithFilePath(""),
		logger.WithCaller(true),
		logger.WithCallerSkip(1),
		logger.WithTraffic(true),
	)
	return func(t *testing.T) {
		// Traffic writes asynchronously via a lumberjack sync; give it a beat
		// to flush before the process exits.
		t.Logf("teardown")
	}
}

// TestChain 验证 DefaultChain 端到端:请求经 [ChainContext] 注入后,
// 链正常执行 handler 并返回响应;request ID 被回写头。
func TestChain(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/test", func(c *gin.Context) {
		handler := RpcHandler(func(ctx context.Context, req any) (resp any, err error) {
			logger.FromContext(ctx).Infof("handle request")
			return gin.H{"hello": "world"}, nil
		})

		ctx := ChainContext(c, "test")
		resp, err := DefaultChain().Intercept(ctx, nil, handler)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	body := []byte(`title=test`)
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Request-Id", "123456")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	t.Logf("status: %d", w.Code)
	t.Logf("response content: %s", w.Body.String())
	// request ID 应被回写(沿用传入的 123456)。
	assert.Equal(t, "123456", w.Header().Get("X-Request-Id"))
}

// TestChainPanicRecover 验证 PanicRecover 兜住 panic 并翻译成 500 error。
func TestChainPanicRecover(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/boom", func(c *gin.Context) {
		handler := RpcHandler(func(ctx context.Context, req any) (resp any, err error) {
			panic("something went wrong")
		})

		ctx := ChainContext(c, "boom")
		_, err := DefaultChain().Intercept(ctx, nil, handler)
		assert.Error(t, err)
		ErrorResponse(c, err)
	})

	req, _ := http.NewRequest("POST", "/boom", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	t.Logf("body: %s", w.Body.String())
}

// TestChainEmpty 直接调用 handler(空链)。
func TestChainEmpty(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	gin.SetMode(gin.TestMode)

	called := false
	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		called = true
		return gin.H{"ok": true}, nil
	})

	resp, err := NewChain().Intercept(context.Background(), nil, h)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, gin.H{"ok": true}, resp)
}

// TestChainWith 追加一个内层拦截器并验证执行顺序(外→内→外)。
func TestChainWith(t *testing.T) {
	teardown := setupLogger(t)
	defer teardown(t)

	var order []string
	outer := Interceptor(func(ctx context.Context, req any, next RpcHandler) (any, error) {
		order = append(order, "outer-in")
		resp, err := next(ctx, req)
		order = append(order, "outer-out")
		return resp, err
	})
	inner := Interceptor(func(ctx context.Context, req any, next RpcHandler) (any, error) {
		order = append(order, "inner-in")
		resp, err := next(ctx, req)
		order = append(order, "inner-out")
		return resp, err
	})
	h := RpcHandler(func(ctx context.Context, req any) (any, error) {
		order = append(order, "handler")
		return nil, nil
	})

	_, _ = NewChain(outer, inner).Intercept(context.Background(), nil, h)
	assert.Equal(t, []string{"outer-in", "inner-in", "handler", "inner-out", "outer-out"}, order)
}
