package ginext

import (
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
		username, _ := c.Get("username")
		c.JSON(http.StatusOK, gin.H{"message": "Hello " + username.(string)})
	})

	validToken, err := GenerateToken("testuser", time.Now().Add(5*time.Minute))
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
			expectedBody: `{"error":"Missing token"}`,
		},
		{
			name:         "Invalid Token",
			token:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzE4NjM4NjIzfQ.3jkMyPp2j7-3EFsLBmMRmTY15JVqmMo8kZGySd7gr-U",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Invalid token"}`,
		},
		{
			name:         "Valid Token",
			token:        validToken,
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Hello testuser"}`,
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
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}
