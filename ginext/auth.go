package ginext

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth is the interface that wraps the Auth method.
//
//go:generate mockery --name Auth --filename auth_mock.go --inpackage
type Auth interface {
	// VerifyUser returns true if the username and password are valid
	VerifyUser(ctx context.Context, username, password string) (bool, error)
	// VerifyToken returns true if the token is valid
	// token is a jwt token
	VerifyToken(ctx context.Context, token string) (bool, error)
	// VerifyAuthentication returns true if the token is valid
	// authentication is Authentication header
	VerifyAuthentication(ctx context.Context, authentication string) (bool, error)
	// VerifyCookie returns true if the cookie is valid
	VerifyCookie(ctx context.Context, cookie string) (bool, error)
	// RefreshToken refreshes the token
	RefreshToken(ctx context.Context, token string) (string, error)
	// RefreshCookie refreshes the cookie
	RefreshCookie(ctx context.Context, cookie string) (string, error)
}

type UserExtractor func(c *gin.Context) (username, password string)

type GinAuth struct {
	auth             Auth
	whitelist        map[string]struct{}
	prefixWhitelist  []string
	loginUriHandlers map[string]UserExtractor
}

func NewGinAuth(
	auth Auth,
	whitelist []string,
	prefixWhitelist []string,
) *GinAuth {
	whitelistMap := make(map[string]struct{}, len(whitelist))
	for _, v := range whitelist {
		whitelistMap[v] = struct{}{}
	}

	return &GinAuth{
		auth:            auth,
		whitelist:       whitelistMap,
		prefixWhitelist: prefixWhitelist,
	}
}

func (ga *GinAuth) Apply() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			ctx    = c.Request.Context()
			reqUrl = c.Request.URL.Path
			query  = c.Request.URL.RawQuery
		)

		// if match whitelist, skip auth
		if _, ok := ga.whitelist[reqUrl]; ok {
			c.Next()
			return
		}

		// if reqUrl start with prefix, skip auth
		for _, prefix := range ga.prefixWhitelist {
			if prefix != "" && strings.HasPrefix(reqUrl, prefix) {
				c.Next()
				return
			}
		}

		// if reqUrl match loginUriHandlers, extract username and password
		if extractor, ok := ga.loginUriHandlers[reqUrl]; ok {
			username, password := extractor(c)
			if username == "" || password == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
				c.Abort()
				return
			}
			if _, err := ga.auth.VerifyUser(ctx, username, password); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
				c.Abort()
				return
			}

			// refresh cookie
			cookie, err := ga.auth.RefreshCookie(ctx, "")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
				c.Abort()
				return
			}

			// set cookie
			c.SetCookie("token", cookie, 3600, "/", "", false, true)

			c.Next()
			return
		}

	}
}
