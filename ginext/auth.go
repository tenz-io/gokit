package ginext

import (
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/errcode"
	"github.com/tenz-io/gokit/logger"
)

var (
	jwtKey = []byte("my_secret_key")
)

const (
	RoleAnonymous = 1
	RoleUser      = 2
	RoleAdmin     = 4
)

type Claims struct {
	Userid int64 `json:"userid"`
	Role   int32 `json:"role"`
	jwt.StandardClaims
}

func InitJWT(secretKey string) {
	jwtKey = []byte(secretKey)
}

func Authenticate(role int32) func(c *gin.Context) {
	if role == 0 || role == RoleAnonymous {
		// skip authentication
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		var (
			le = logger.FromContext(c.Request.Context())
		)
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			le.Warnf("missing token")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "missing token"))
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

			if e := new(jwt.ValidationError); errors.As(err, &e) {
				if e.Errors&jwt.ValidationErrorMalformed != 0 {
					ErrorResponse(c, errcode.BadRequest(http.StatusBadRequest, "bad token"))
					return
				}
				if e.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token"))
					return
				}
				ErrorResponse(c, errcode.BadRequest(http.StatusBadRequest, "bad request"))
				return
			}

			ErrorResponse(c, errcode.BadRequest(http.StatusBadRequest, "bad request"))
			return
		}

		if !token.Valid {
			le.Warnf("invalid token")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "invalid token"))
			return
		}

		// check role
		if role&claims.Role == 0 {
			le.Warnf("role not match")
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "role not match"))
			return
		}

		le.Debugf("authenticated user: %d", claims.Userid)
		c.Set("userid", claims.Userid)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func GenerateToken(userid int64, role int32, expiredAt time.Time) (string, error) {
	claims := &Claims{
		Userid: userid,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredAt.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
