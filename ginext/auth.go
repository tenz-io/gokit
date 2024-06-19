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

const (
	RoleAnonymous = 0
	RoleAdmin     = 1
	RoleUser      = 2
)

const (
	CookieTokenName  = "token"
	ExpiresInMinutes = 15
)

type Claims struct {
	Userid int64 `json:"userid"`
	Role   int32 `json:"role"`
	jwt.RegisteredClaims
}

func InitJWT(secretKey string) {
	jwtKey = []byte(secretKey)
}

func Authenticate(role int32, cookie bool) func(c *gin.Context) {
	if cookie {
		return AuthenticateCookie(role)
	} else {
		return AuthenticateRest(role)
	}
}

// AuthenticateRest is a middleware to authenticate user by token in Authorization header
func AuthenticateRest(role int32) func(c *gin.Context) {
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

		claims := Claims{}
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
			// Ensure the signing method is HMAC and the key is correct
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtKey, nil
		})
		if err != nil {
			le.Warnf("error parsing token: %v", err)
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token"))
				return
			}

			ErrorResponse(c, errcode.BadRequest(http.StatusBadRequest, "bad token in request"))
			return
		}

		le = le.WithFields(logger.Fields{
			"userid":    claims.Userid,
			"user_role": claims.Role,
		})

		if !token.Valid {
			le.Warnf("invalid token")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token"))
			return
		}

		// check admin
		if role == RoleAdmin && claims.Role != RoleAdmin {
			le.Warnf("require admin role")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "role not match"))
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
		ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "role not match"))
		return
	}
}

// AuthenticateCookie is a middleware to authenticate user by token in cookie
func AuthenticateCookie(role int32) func(c *gin.Context) {
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

		claims := Claims{}
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
			// Ensure the signing method is HMAC and the key is correct
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtKey, nil
		})
		if err != nil {
			le.Warnf("error parsing token: %v", err)
			if errors.Is(err, jwt.ErrSignatureInvalid) {
				ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token"))
				return
			}

			ErrorResponse(c, errcode.BadRequest(http.StatusBadRequest, "bad token in request"))
			return
		}

		le = le.WithFields(logger.Fields{
			"userid":    claims.Userid,
			"user_role": claims.Role,
		})

		if !token.Valid {
			le.Warnf("invalid token")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token"))
			return
		}

		// check admin
		if role == RoleAdmin && claims.Role != RoleAdmin {
			le.Warnf("require admin role")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "role not match"))
			return
		}

		// admin can access all resources
		if claims.Role == RoleAdmin || role&claims.Role > 0 {
			le.Debugf("authenticated")
			c.Set("userid", claims.Userid)
			c.Set("role", claims.Role)

			// refresh token
			expiredAt := time.Now().Add(ExpiresInMinutes * time.Minute)
			newToken, err := GenerateToken(claims.Userid, claims.Role, expiredAt)
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
		ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "role not match"))
		return
	}
}

func GenerateToken(userid int64, role int32, expiredAt time.Time) (string, error) {
	claims := &Claims{
		Userid: userid,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
