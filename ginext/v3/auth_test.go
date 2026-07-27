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

// newTestAuth 构造一个带固定密钥(>=32 字节)的 Auth,供测试用。
func newTestAuth() *Auth {
	auth, err := NewAuth("test_secret_key_at_least_32_bytes_long!!")
	if err != nil {
		panic(err)
	}
	return auth
}

func TestRequireRoleHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := newTestAuth()

	router := gin.New()
	router.GET("/protected", auth.RequireRoleHeader(RoleUser), func(c *gin.Context) {
		userid, _ := c.Get("user_id")
		roles, _ := c.Get("roles")
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user_id: %v, roles: %v", userid, roles)})
	})

	validToken, err := auth.GenerateToken(123, []string{RoleUser}, TokenTypeAccess, time.Now().Add(5*time.Minute))
	assert.NoError(t, err)

	// 缺失 requiredRoles(空) → 匿名放行,不应 401。
	anonToken, err := auth.GenerateToken(123, []string{RoleUser}, TokenTypeAccess, time.Now().Add(5*time.Minute))
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
			expectedBody: `{"code":401,"message":"unauthorized","data":{}}`,
		},
		{
			name:         "Invalid Token",
			token:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNzE4NjM4NjIzfQ.3jkMyPp2j7-3EFsLBmMRmTY15JVqmMo8kZGySd7gr-U",
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":401,"message":"unauthorized","data":{}}`,
		},
		{
			name:         "Valid Token",
			token:        validToken,
			expectedCode: http.StatusOK,
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
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}

	// 显式覆盖:空 requiredRoles 放行匿名。
	t.Run("empty required roles allows anonymous", func(t *testing.T) {
		r := gin.New()
		r.GET("/open", auth.RequireRoleHeader(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})
		req, _ := http.NewRequest(http.MethodGet, "/open", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// refresh token 不应通过(access-only)。
	t.Run("refresh token rejected", func(t *testing.T) {
		refresh, err := auth.GenerateToken(123, []string{RoleUser}, TokenTypeRefresh, time.Now().Add(5*time.Minute))
		assert.NoError(t, err)
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", refresh)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// 角色不匹配:user token 不能访问 admin。
	t.Run("role mismatch rejected", func(t *testing.T) {
		adminRouter := gin.New()
		adminRouter.GET("/admin", auth.RequireRoleHeader(RoleAdmin), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin"})
		})
		req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", validToken) // user token
		w := httptest.NewRecorder()
		adminRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	_ = anonToken // silence unused if branch above not exercised
}

func TestIsAuthedHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := newTestAuth()

	router := gin.New()
	router.GET("/protected", func(c *gin.Context) {
		if !auth.IsAuthedHeader(c, RoleUser) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}
		userid, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user_id: %v", userid)})
	})

	validToken, err := auth.GenerateToken(123, []string{RoleUser}, TokenTypeAccess, time.Now().Add(5*time.Minute))
	assert.NoError(t, err)

	tests := []struct {
		name         string
		token        string
		expectedCode int
	}{
		{name: "Missing Token", token: "", expectedCode: http.StatusUnauthorized},
		{name: "Invalid Token", token: "invalid", expectedCode: http.StatusUnauthorized},
		{name: "Valid Token", token: validToken, expectedCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestRequireRoleCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth := newTestAuth()

	router := gin.New()
	router.GET("/protected", auth.RequireRoleCookie(RoleUser), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	validToken, err := auth.GenerateToken(123, []string{RoleUser}, TokenTypeAccess, time.Now().Add(5*time.Minute))
	assert.NoError(t, err)

	tests := []struct {
		name         string
		cookie       string
		expectedCode int
	}{
		{name: "Missing Cookie", cookie: "", expectedCode: http.StatusUnauthorized},
		{name: "Invalid Cookie", cookie: "invalid", expectedCode: http.StatusUnauthorized},
		{name: "Valid Cookie", cookie: validToken, expectedCode: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			if tt.cookie != "" {
				req.Header.Set("Cookie", fmt.Sprintf("%s=%s", CookieTokenName, tt.cookie))
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}

	// 鉴权成功的响应应回写一个新的(续期)cookie。
	t.Run("valid cookie is refreshed", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Cookie", fmt.Sprintf("%s=%s", CookieTokenName, validToken))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		setCookie := w.Header().Get("Set-Cookie")
		assert.Contains(t, setCookie, CookieTokenName+"=", "expect refreshed token cookie")
	})
}

func TestIsUnauthorizedError(t *testing.T) {
	assert.True(t, IsUnauthorizedError(ErrUnauthorized))
	assert.True(t, IsUnauthorizedError(ErrInvalidToken))
	assert.False(t, IsUnauthorizedError(nil))
	assert.False(t, IsUnauthorizedError(fmt.Errorf("other")))
}
