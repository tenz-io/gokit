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

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func InitJWT(secretKey string) {
	jwtKey = []byte(secretKey)
}

func Authenticate(c *gin.Context) {
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

	le.Debugf("authenticated user: %s", claims.Username)
	c.Set("username", claims.Username)
	c.Set("role", claims.Role)
	c.Next()
}

func GenerateToken(username, role string, expiredAt time.Time) (string, error) {
	claims := &Claims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredAt.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
