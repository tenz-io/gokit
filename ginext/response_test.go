package ginext

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockFileResponse struct{}

func (mfr MockFileResponse) GetFile() []byte {
	return []byte("file content")
}

func TestResponse(t *testing.T) {
	// Initialize Gin Engine in test mode
	gin.SetMode(gin.TestMode)

	// Define test cases
	tests := []struct {
		name       string
		data       any
		expectBody string
		expectCode int
	}{
		{
			name:       "nil data",
			data:       nil,
			expectBody: "{\"code\":0,\"data\":{},\"message\":\"success\"}",
			expectCode: http.StatusOK,
		},
		{
			name:       "normal data",
			data:       gin.H{"hello": "world"},
			expectBody: "{\"code\":0,\"data\":{\"hello\":\"world\"},\"message\":\"success\"}",
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
			// Create a response recorder
			w := httptest.NewRecorder()

			// Create a new context with the response recorder
			c, _ := gin.CreateTestContext(w)

			// Call the function to test
			Response(c, tt.data)

			// Assert HTTP status code
			assert.Equal(t, tt.expectCode, w.Code)

			// Assert body content
			if tt.name == "file data" {
				assert.Equal(t, tt.expectBody, string(w.Body.Bytes()))
				// assert content type
				assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain"))
			} else {
				assert.JSONEq(t, tt.expectBody, w.Body.String())
			}
		})
	}
}
