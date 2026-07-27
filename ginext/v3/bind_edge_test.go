package ginext

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/tenz-io/gokit/annotation/v3"
	"github.com/tenz-io/gokit/ginext/v3/errcode"
)

// TestBindAndValidate_MissingRequiredQuery 验证缺失必填 query 字段转 400。
func TestBindAndValidate_MissingRequiredQuery(t *testing.T) {
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

	// 缺 page/page_size(虽有 default,但 required 校验应通过 default 满足);
	// 这里给 page 一个非法值触发 query required 的另一条分支。
	body := []byte(`title=t`)
	req, _ := http.NewRequest("POST", "/v1/1/articles?page=0", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// page=0 违反 gt=0 → 校验失败 → 400
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindAndValidate_InvalidURIType 验证 uri 参数类型转换失败转 400。
func TestBindAndValidate_InvalidURIType(t *testing.T) {
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

	// userid 应为 int64,传非数字 → SetString 失败 → 400
	body := []byte(`title=t`)
	req, _ := http.NewRequest("POST", "/v1/abc/articles?page=1&page_size=10", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindAndValidate_InvalidQueryType 验证 query 参数类型转换失败转 400。
func TestBindAndValidate_InvalidQueryType(t *testing.T) {
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

	// page 应为 int32,传非数字 → SetString 失败 → 400
	body := []byte(`title=t`)
	req, _ := http.NewRequest("POST", "/v1/1/articles?page=abc&page_size=10", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindAndValidate_FormMethodNotAllowed 验证非 POST/PUT 的 form 请求转 400。
func TestBindAndValidate_FormMethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/form", func(c *gin.Context) {
		var in TestRequest
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &in)
	})

	req, _ := http.NewRequest(http.MethodGet, "/form?title=t&page=1&page_size=10", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindAndValidate_MultipartMethodNotAllowed 验证非 POST/PUT 的 multipart 请求转 400。
func TestBindAndValidate_MultipartMethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/upload", func(c *gin.Context) {
		var in TestFileRequest
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &in)
	})

	req, _ := http.NewRequest(http.MethodGet, "/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBindAndValidate_MultipartNoFileField 验证 multipart 无文件字段转 400。
func TestBindAndValidate_MultipartNoFileField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 用一个没有 bind:"file" 字段的结构体,Content-Type 是 multipart。
	router.POST("/upload", func(c *gin.Context) {
		var in struct {
			Title string `bind:"form,name=title" validate:"required"`
		}
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, &in)
	})

	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	w.WriteField("title", "t")
	w.Close()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestBindAndValidate_BindErrorIsErrcode 验证所有绑定错误都被 warpError 包成 errcode(可被 ErrorResponse 映射)。
func TestBindAndValidate_BindErrorIsErrcode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.POST("/v1/:userid/articles", func(c *gin.Context) {
		var in TestRequest
		err := BindAndValidate(c, &in)
		assert.Error(t, err)
		e, ok := errcode.FromError(err)
		assert.True(t, ok, "bind error must be errcode.Error")
		if ok {
			assert.Equal(t, http.StatusBadRequest, e.Status)
			assert.Equal(t, http.StatusBadRequest, e.Code)
		}
		Response(c, nil)
	})

	body := []byte(`title=`)
	req, _ := http.NewRequest("POST", "/v1/abc/articles?page=x", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code) // handler 手动 Response
}

// TestTryBindForm_NotFormContentType 验证非 form Content-Type 时 tryBindForm 返回 has=false。
func TestTryBindForm_NotFormContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/json"}},
	}
	has, err := tryBindForm(c, &TestRequest{})
	assert.False(t, has)
	assert.NoError(t, err)
}

// TestTryBindJSON_NotJSONContentType 验证非 json Content-Type 时 tryBindJSON 返回 has=false。
func TestTryBindJSON_NotJSONContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Method: http.MethodPost,
		Header: http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
	}
	has, err := tryBindJSON(c, &TestRequest{})
	assert.False(t, has)
	assert.NoError(t, err)
}

// TestTryBindJSON_GETSkipped 验证 GET 请求即使 Content-Type 是 json 也不绑定 body。
func TestTryBindJSON_GETSkipped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/x", bytes.NewReader([]byte(`{"title":"t"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	has, err := tryBindJSON(c, &TestRequest{})
	assert.False(t, has)
	assert.NoError(t, err)
}

// TestReadFileAndSetField_EmptyFile 验证空文件(0 字节)不报错、不设置字段。
func TestReadFileAndSetField_EmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// 构造一个 multipart 请求,文件内容为空。
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("txt_file", "empty.txt")
	assert.NoError(t, err)
	fw.Write([]byte{})
	w.Close()

	c.Request, _ = http.NewRequest("POST", "/upload", body)
	c.Request.Header.Set("Content-Type", w.FormDataContentType())

	in := &TestFileRequest{}
	root, _ := rootValue(in)
	plan, _ := annotation.PlanFor(in)
	for _, f := range plan.FieldsBySource("file") {
		e := readFileAndSetField(c, root, f)
		assert.NoError(t, e, "empty file should not error")
	}
}

// TestBindAndValidate_DefaultsApplied 验证 default 标签在可选字段缺失时填充。
// 注意:default 不覆盖 required —— 一个 required 字段即便有 default,在绑定
// 来源取不到值时仍会报 "is required"。故这里用一个带 default 但无 required 的
// 可选字段验证默认值填充。
func TestBindAndValidate_DefaultsApplied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	type Req struct {
		Name   string `bind:"form,name=name" validate:"required,min_len=1"`
		Offset int32  `bind:"query,name=offset" default:"7" validate:"gte=0"`
	}

	router.POST("/item", func(c *gin.Context) {
		var in Req
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, gin.H{"offset": in.Offset})
	})

	body := []byte(`name=t`)

	// 缺失 offset → default 填 7。
	req, _ := http.NewRequest("POST", "/item", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"code":0,"message":"success","data":{"offset":7}}`, w.Body.String())

	// 显式传值时 default 不覆盖。
	req2, _ := http.NewRequest("POST", "/item?offset=42", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.JSONEq(t, `{"code":0,"message":"success","data":{"offset":42}}`, w2.Body.String())
}
