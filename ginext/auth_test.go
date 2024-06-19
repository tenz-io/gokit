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
	router.GET("/protected", Authenticate(RoleUser, AuthTypeRest), func(c *gin.Context) {
		userid := c.GetInt64("userid")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("userid: %d, role: %d", userid, role)})
	})

	validToken, err := GenerateToken(123, RoleUser, TokenTypeAccess, time.Now().Add(5*time.Minute))
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
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"bad token in request","data":{}}`,
		},
		{
			name:         "Valid Token",
			token:        validToken,
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"userid: 123, role: 2"}`,
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
