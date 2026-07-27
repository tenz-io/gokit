package ginext

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestBind_JSONDoesNotOverlayURI 验证 JSON body 不得覆盖 URI 路径参数(finding 1)。
// /user/1 配 {"id":2}:ID 应保持 1,不被 body 改写。
func TestBind_JSONDoesNotOverlayURI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	type Req struct {
		ID   int64  `json:"id" bind:"uri,name=id" validate:"required,gt=0"`
		Name string `json:"name" validate:"required"`
	}

	router.POST("/user/:id", func(c *gin.Context) {
		var in Req
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, gin.H{"id": in.ID, "name": in.Name})
	})

	// body 试图把 id 改成 2,但 URI 声明了 id=1 —— 应以 URI 为准。
	body := []byte(`{"id":2,"name":"bob"}`)
	req, _ := http.NewRequest("POST", "/user/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"code":0,"message":"success","data":{"id":1,"name":"bob"}}`,
		w.Body.String(), "URI id must not be overwritten by JSON body")
}

// TestBind_JSONDoesNotOverlayQuery 验证 JSON body 不得覆盖 query 参数(finding 1)。
func TestBind_JSONDoesNotOverlayQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	type Req struct {
		Name   string `json:"name" validate:"required"`
		Offset int32  `json:"offset" bind:"query,name=offset" validate:"gte=0"`
	}

	router.POST("/items", func(c *gin.Context) {
		var in Req
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, gin.H{"offset": in.Offset, "name": in.Name})
	})

	// ?offset=5 但 body 给 offset=999 —— 应以 query 为准。
	body := []byte(`{"name":"x","offset":999}`)
	req, _ := http.NewRequest("POST", "/items?offset=5", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"code":0,"message":"success","data":{"offset":5,"name":"x"}}`,
		w.Body.String(), "query offset must not be overwritten by JSON body")
}

// TestBind_JSONDoesNotOverlayHeader 验证 JSON body 不得覆盖 header(finding 1)。
func TestBind_JSONDoesNotOverlayHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	type Req struct {
		Name string `json:"name" validate:"required"`
		Auth string `json:"auth" bind:"header,name=Authorization" validate:"required"`
	}

	router.POST("/x", func(c *gin.Context) {
		var in Req
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, gin.H{"auth": in.Auth, "name": in.Name})
	})

	body := []byte(`{"name":"x","auth":"EVIL"}`)
	req, _ := http.NewRequest("POST", "/x", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer real-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"code":0,"message":"success","data":{"auth":"Bearer real-token","name":"x"}}`,
		w.Body.String(), "header must not be overwritten by JSON body")
}

// TestBind_OptionalFormTypeErrorRejected 验证可选 form 字段的类型转换错误
// 不再被吞,而是返回 400(finding 6)。
func TestBind_OptionalFormTypeErrorRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	type Req struct {
		Name string `bind:"form,name=name" validate:"required"`
		Age  int32  `bind:"form,name=age" validate:"gte=0"` // 可选(无 required)
	}

	router.POST("/person", func(c *gin.Context) {
		var in Req
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, gin.H{"age": in.Age, "name": in.Name})
	})

	// age=abc:字段出现但类型错误 → 必须 400,不能静默置零。
	body := []byte(`name=bob&age=abc`)
	req, _ := http.NewRequest("POST", "/person", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// age 缺失:可选,应成功置零。
	body2 := []byte(`name=bob`)
	req2, _ := http.NewRequest("POST", "/person", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

// TestBind_OptionalFormTypeErrorRejectedJSON 验证可选字段在 JSON 请求里类型错误也 400。
func TestBind_OptionalFormTypeErrorRejectedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	type Req struct {
		Name string `json:"name" validate:"required"`
		Age  int32  `json:"age" validate:"gte=0"`
	}

	router.POST("/person", func(c *gin.Context) {
		var in Req
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, gin.H{"age": in.Age, "name": in.Name})
	})

	// age 为字符串 → json unmarshal 失败 → 400。
	body := []byte(`{"name":"bob","age":"abc"}`)
	req, _ := http.NewRequest("POST", "/person", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBind_OversizeBodyRejected 验证请求体超过上限时返回 413(finding 8)。
// 用临时下调 SetMaxBodyBytes 避免 10MB 的测试开销,测后恢复默认。
func TestBind_OversizeBodyRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	type Req struct {
		Name string `json:"name" validate:"required"`
	}

	router.POST("/person", func(c *gin.Context) {
		var in Req
		if err := BindAndValidate(c, &in); err != nil {
			ErrorResponse(c, err)
			return
		}
		Response(c, gin.H{"name": in.Name})
	})

	// 临时设为 16 字节上限,测后恢复。
	SetMaxBodyBytes(16)
	defer SetMaxBodyBytes(0) // 0 → 恢复默认

	// 32 字节 JSON 超过 16 → 413。
	body := bytes.Repeat([]byte("a"), 32)
	req, _ := http.NewRequest("POST", "/person", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)

	// 小于上限的合法请求仍成功(16 字节以内)。
	body2 := []byte(`{"name":"ab"}`) // 13 字节
	req2, _ := http.NewRequest("POST", "/person", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
