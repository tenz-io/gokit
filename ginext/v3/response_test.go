package ginext

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/tenz-io/gokit/ginext/v3/errcode"
)

type MockFileResponse struct{}

func (mfr MockFileResponse) GetFile() []byte {
	return []byte("file content")
}

func TestResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		data       any
		expectBody string
		expectCode int
	}{
		{
			name:       "nil data",
			data:       nil,
			expectBody: `{"code":0,"data":{},"message":"success"}`,
			expectCode: http.StatusOK,
		},
		{
			name:       "normal data",
			data:       gin.H{"hello": "world"},
			expectBody: `{"code":0,"data":{"hello":"world"},"message":"success"}`,
			expectCode: http.StatusOK,
		},
		{
			name:       "file data",
			data:       MockFileResponse{},
			expectBody: "file content",
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Response(c, tt.data)

			assert.Equal(t, tt.expectCode, w.Code)

			if tt.name == "file data" {
				assert.Equal(t, tt.expectBody, string(w.Body.Bytes()))
				assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain"))
			} else {
				assert.JSONEq(t, tt.expectBody, w.Body.String())
			}
		})
	}
}

func TestResponseStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("created status writes HTTP 201 with frame", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ResponseStatus(c, http.StatusCreated, gin.H{"id": 1})

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.JSONEq(t,
			`{"code":0,"data":{"id":1},"message":"success"}`,
			w.Body.String())
	})

	t.Run("file data ignores status, stays 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ResponseStatus(c, http.StatusCreated, MockFileResponse{})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "file content", string(w.Body.Bytes()))
		assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain"))
	})
}

func TestErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("errcode error maps status and code", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ErrorResponse(c, errcode.NotFound(http.StatusNotFound, "not found"))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t,
			`{"code":404,"message":"not found","data":{}}`,
			w.Body.String())
	})

	t.Run("plain error falls back to 500 with generic message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ErrorResponse(c, errors.New("boom: connection refused to db.internal:5432"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// 原始错误细节不得外泄,只回固定通用消息。
		body := w.Body.String()
		assert.NotContains(t, body, "db.internal", "raw error must not leak")
		assert.NotContains(t, body, "boom", "raw error must not leak")
		assert.Contains(t, body, "internal server error")
	})
}
