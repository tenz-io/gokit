package ginext

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/tenz-io/gokit/ginext/v3/errcode"
)

// TestValidate 验证 Validate 仅校验不绑定。
func TestValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid struct passes", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		in := &TestRequest{Title: "ok", UserID: 1, Page: 1, PageSize: 10}
		err := Validate(c, in)
		assert.NoError(t, err)
	})

	t.Run("invalid struct fails with 400 errcode", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		in := &TestRequest{Title: "", UserID: 0} // title required, userid required gt=0
		err := Validate(c, in)
		assert.Error(t, err)
		e, ok := errcode.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, e.Status)
	})
}

// TestWarpError 单测错误包装:校验错误、JSON 错误、普通错误三路。
func TestWarpError(t *testing.T) {
	t.Run("nil is nil", func(t *testing.T) {
		assert.NoError(t, warpError(nil))
	})

	t.Run("plain error wraps to 400", func(t *testing.T) {
		err := warpError(assert.AnError)
		e, ok := errcode.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, e.Status)
	})
}

// TestBindAndValidate_ValidationFailure 验证校验失败转 400 并经 ErrorResponse 渲染。
func TestBindAndValidate_ValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:userid/articles", func(c *gin.Context) {
		var in TestRequest
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &in)
	})

	// title 缺失(校验 required),page/page_size 缺省填 1/10,userid 来自路由。
	body := []byte(`title=`)
	req, _ := http.NewRequest("POST", "/v1/1/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	t.Logf("body: %s", w.Body.String())
	// 响应体应是 {code,message,data} 框架。
	var frame ResponseFrame
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &frame))
	assert.NotZero(t, frame.Code)
	assert.NotEmpty(t, frame.Message)
}

// TestBindAndValidate_MalformedJSON 验证畸形 JSON 体转 400。
func TestBindAndValidate_MalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:userid/articles", func(c *gin.Context) {
		var in TestRequest
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &in)
	})

	body := []byte(`{not json`)
	req, _ := http.NewRequest("POST", "/v1/1/articles?page=1&page_size=10", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindAndValidate_MissingRequiredURI 验证缺失必填 uri 参数转 400。
func TestBindAndValidate_MissingRequiredURI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 路由不含 :userid,故 uri 绑定取不到值 → required 失败。
	router.POST("/v1/articles", func(c *gin.Context) {
		var in TestRequest
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &in)
	})

	body := []byte(`title=ok`)
	req, _ := http.NewRequest("POST", "/v1/articles?page=1&page_size=10", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindAndValidate_NonNilError 验证 BindAndValidate 返回的错误确实是 errcode(可被 ErrorResponse 映射)。
func TestBindAndValidate_NonNilError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/x", func(c *gin.Context) {
		var in TestRequest
		err := BindAndValidate(c, &in)
		assert.Error(t, err)
		_, ok := errcode.FromError(err)
		assert.True(t, ok, "bind error must be an errcode.Error")
		Response(c, nil)
	})

	body := []byte(`title=`)
	req, _ := http.NewRequest("POST", "/x", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code) // handler 里手动 Response 了
}

// TestRootValue_ErrorsOnNonPointer 验证 rootValue 拒绝非指针。
func TestRootValue_ErrorsOnNonPointer(t *testing.T) {
	_, err := rootValue("not a pointer")
	assert.Error(t, err)

	_, err = rootValue(nil)
	assert.Error(t, err)
}
