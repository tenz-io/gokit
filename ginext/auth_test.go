package ginext

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/protected", Authenticate, func(c *gin.Context) {
		username := c.GetString("username")
		role := c.GetString("role")
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Hello %s, you are %s", username, role)})
	})

	validToken, err := GenerateToken("testuser", "admin", time.Now().Add(5*time.Minute))
	assert.NoError(t, err)

	tests := []struct {
		name         string
		token        string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Missing Token",
			token:        "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":401,"message":"missing token","data":{}}`,
		},
		{
			name:         "Invalid Token",
			token:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzE4NjM4NjIzfQ.3jkMyPp2j7-3EFsLBmMRmTY15JVqmMo8kZGySd7gr-U",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":401,"message":"invalid token","data":{}}`,
		},
		{
			name:         "Valid Token",
			token:        validToken,
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Hello testuser, you are admin"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			t.Logf("token: %s", tt.token)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			t.Logf("body: %s", w.Body.String())
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
