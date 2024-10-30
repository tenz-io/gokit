package ginext

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/tenz-io/gokit/ginext/errcode"
	"github.com/tenz-io/gokit/logger"
)

var (
	jwtKey = []byte("my_secret_key")
)

type (
	// RoleType is the role type for user, including:
	// 0: anonymous
	// 1: admin
	// 2: user
	RoleType = int32
	// AuthType is the auth type, including:
	// 0: web page
	// 1: restful api
	AuthType = int32
	// TokenType is the token type, including:
	// 0: access token
	// 1: refresh token
	TokenType = int32
)

const (
	RoleAnonymous RoleType = 0
	RoleAdmin     RoleType = 1
	RoleUser      RoleType = 2
)

const (
	// AuthTypeWeb is the auth type for web page, @see genproto/api/custom/common/authz.proto
	AuthTypeWeb  AuthType = 0
	AuthTypeRest AuthType = 1
)

const (
	TokenTypeAccess  TokenType = 0
	TokenTypeRefresh TokenType = 1
)

const (
	CookieTokenName  = "token"
	ExpiresInMinutes = 15
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrUnauthorized = errors.New("unauthorized")
)

type Claims struct {
	Userid int64 `json:"userid"`
	Role   int32 `json:"role"` // 0 anonymous, 1 admin, 2 user
	Type   int32 `json:"type"` // 0 access, 1 refresh
	jwt.RegisteredClaims
}

func InitJWT(secretKey string) {
	jwtKey = []byte(secretKey)
}

func Authenticate(role RoleType, authType AuthType) func(c *gin.Context) {
	if role == RoleAnonymous {
		// skip authentication
		return func(c *gin.Context) {
			c.Next()
		}
	}

	switch authType {
	case AuthTypeWeb:
		return AuthenticateCookie(role)
	case AuthTypeRest:
		return AuthenticateRest(role)
	default:
		return func(c *gin.Context) {
			ErrorResponse(c, errcode.BadRequest(http.StatusInternalServerError, "bad auth type"))
		}
	}
}

// AuthenticateRest is a middleware to authenticate user by token in Authorization header
func AuthenticateRest(role RoleType) func(c *gin.Context) {
	if role == RoleAnonymous {
		// skip authentication
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		var (
			le = logger.FromContext(c.Request.Context()).WithFields(logger.Fields{
				"required_role": role,
			})
		)
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			le.Warnf("missing token")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "missing token"))
			return
		}

		// remove "Bearer " prefix
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimSpace(tokenString[7:])
		}

		claims, err := VerifyToken(tokenString)
		if err != nil {
			le.Warnf("error parsing token: %v", err)
			if IsUnauthorizedError(err) {
				ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token"))
				return
			}

			ErrorResponse(c, errcode.BadRequest(http.StatusBadRequest, "bad token in request"))
			return
		}

		le = le.WithFields(logger.Fields{
			"userid":     claims.Userid,
			"user_role":  claims.Role,
			"token_type": claims.Type,
		})

		// token type check
		if claims.Type != TokenTypeAccess {
			le.Warnf("invalid token type")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token type"))
			return
		}

		// check admin
		if role == RoleAdmin && claims.Role != RoleAdmin {
			le.Warnf("require admin role")
			ErrorResponse(c, errcode.Unauthorized(http.StatusForbidden, "forbidden"))
			return
		}

		// admin can access all resources
		if claims.Role == RoleAdmin || role&claims.Role > 0 {
			le.Debugf("authenticated")
			c.Set("userid", claims.Userid)
			c.Set("role", claims.Role)
			c.Next()
			return
		}

		le.Warnf("role not match")
		ErrorResponse(c, errcode.Unauthorized(http.StatusForbidden, "forbidden"))
		return
	}
}

// AuthenticateCookie is a middleware to authenticate user by token in cookie
func AuthenticateCookie(role RoleType) func(c *gin.Context) {
	if role == RoleAnonymous {
		// skip authentication
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		var (
			le = logger.FromContext(c.Request.Context()).WithFields(logger.Fields{
				"required_role": role,
			})
		)

		// get token from cookie
		tokenString, err := c.Cookie(CookieTokenName)
		if err != nil {
			le.Warnf("missing token")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "login required"))
			return
		}

		claims, err := VerifyToken(tokenString)
		if err != nil {
			le.Warnf("error parsing token: %v", err)
			if IsUnauthorizedError(err) {
				ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token"))
				return
			}

			ErrorResponse(c, errcode.BadRequest(http.StatusBadRequest, "bad token in request"))
			return
		}

		le = le.WithFields(logger.Fields{
			"userid":     claims.Userid,
			"user_role":  claims.Role,
			"token_type": claims.Type,
		})

		// token type check
		if claims.Type != TokenTypeAccess {
			le.Warnf("invalid token type")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token type"))
			return
		}

		// check admin
		if role == RoleAdmin && claims.Role != RoleAdmin {
			le.Warnf("require admin role")
			ErrorResponse(c, errcode.Unauthorized(http.StatusForbidden, "forbidden"))
			return
		}

		// admin can access all resources
		if claims.Role == RoleAdmin || role&claims.Role > 0 {
			le.Debugf("authenticated")
			c.Set("userid", claims.Userid)
			c.Set("role", claims.Role)

			// refresh token
			expiredAt := time.Now().Add(ExpiresInMinutes * time.Minute)
			newToken, err := GenerateToken(claims.Userid, claims.Role, TokenTypeAccess, expiredAt)
			if err != nil {
				le.Warnf("failed to generate token: %v", err)
				ErrorResponse(c, errcode.InternalServer(http.StatusInternalServerError, "failed to generate token"))
				return
			}
			c.SetCookie("token", newToken, ExpiresInMinutes*60, "/", "", false, true)

			c.Next()
			return
		}

		le.Warnf("role not match")
		ErrorResponse(c, errcode.Unauthorized(http.StatusForbidden, "forbidden"))
		return
	}
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *gin.Context, role RoleType, authType AuthType) bool {
	if role == RoleAnonymous {
		// skip authentication
		return true
	}

	switch authType {
	case AuthTypeWeb:
		return IsAuthenticateCookie(c, role)
	case AuthTypeRest:
		return IsAuthenticateRest(c, role)
	default:
		return false
	}
}

// isTokenAuthenticated checks if the token is valid and the role is matched
func isTokenAuthenticated(c *gin.Context, role RoleType, token string) bool {
	var (
		ctx = c.Request.Context()
		le  = logger.FromContext(ctx).WithFields(logger.Fields{
			"role": role,
		})
	)

	if role == RoleAnonymous {
		le.Debugf("skip authentication")
		return true
	}

	if token == "" {
		le.Warnf("missing token")
		return false
	}

	claims, err := VerifyToken(token)
	if err != nil {
		le.WithError(err).Warnf("error parsing token")
		return false
	}

	if claims.Type != TokenTypeAccess {
		le.Warnf("invalid token type")
		return false
	}

	if role == RoleAdmin && claims.Role != RoleAdmin {
		le.Warnf("require admin role")
		return false
	}

	if claims.Role == RoleAdmin || role&claims.Role > 0 {
		le.Debugf("authenticated")
		c.Set("userid", claims.Userid)
		c.Set("role", claims.Role)
		return true
	}

	le.Debugf("role not match")
	return false
}

// IsAuthenticateRest is a middleware to check if user is authenticated by token in Authorization header
func IsAuthenticateRest(c *gin.Context, role RoleType) bool {
	var (
		ctx = c.Request.Context()
		le  = logger.FromContext(ctx).WithFields(logger.Fields{
			"role": role,
		})
	)

	if role == RoleAnonymous {
		le.Debugf("skip authentication")
		return true
	}

	tokenString := c.GetHeader("Authorization")
	return isTokenAuthenticated(c, role, tokenString)
}

// IsAuthenticateCookie is a middleware to check if user is authenticated by token in cookie
func IsAuthenticateCookie(c *gin.Context, role RoleType) bool {
	var (
		ctx = c.Request.Context()
		le  = logger.FromContext(ctx).WithFields(logger.Fields{
			"role": role,
		})
	)

	if role == RoleAnonymous {
		le.Debugf("skip authentication")
		return true
	}

	tokenString, err := c.Cookie(CookieTokenName)
	if err != nil {
		le.Warnf("missing token")
		return false
	}

	return isTokenAuthenticated(c, role, tokenString)
}

// GenerateToken generates a token with userid, role, token type and expired time
func GenerateToken(userid int64, role RoleType, tokenType TokenType, expiredAt time.Time) (string, error) {
	claims := &Claims{
		Userid: userid,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// VerifyToken verifies the token and returns the claims if valid
// returns error if the token is invalid
func VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, errors.Join(ErrUnauthorized, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil

}

func IsUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInvalidToken)
}
