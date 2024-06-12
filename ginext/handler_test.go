package ginext

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/metadata"
	"github.com/tenz-io/gokit/logger"
)

func TestAllRpcInterceptor(t *testing.T) {

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/test", func(c *gin.Context) {
		handler := RpcHandler(func(ctx context.Context, req any) (resp any, err error) {
			logger.FromContext(ctx).Infof("handle request")
			return gin.H{
				"hello": "world",
			}, nil
		})

		var (
			ctx = c.Request.Context()
			in  TestRequest
		)

		md := metadata.New(c, "test")
		ctx = metadata.WithMetadata(ctx, md)

		resp, err := AllRpcInterceptor.Intercept(ctx, in, handler)
		if err != nil {
			t.Logf("error: %+v", err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	body := []byte(`title=test`)
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Request-ID", "123456")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	t.Logf("status: %d", w.Code)
	respContent := w.Body.String()
	t.Logf("response content: %s", respContent)

}
