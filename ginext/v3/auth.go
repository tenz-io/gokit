package ginext

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/tenz-io/gokit/ginext/v3/errcode"
	"github.com/tenz-io/gokit/logger/v3"
)

// 角色以字符串显式声明,而非 v2 的 bit-mask int。这避免了 "匿名=0 且
// 0 & x 恒为 0" 之类的魔法值陷阱,也让角色集合可读、可扩展。
const (
	// RoleAdmin 是管理员角色。
	RoleAdmin = "admin"
	// RoleUser 是普通用户角色。
	RoleUser = "user"
	// RoleAnonymous 表示匿名(未登录)。当它出现在要求列表或要求列表为空时,
	// 鉴权中间件放行所有请求 —— 与 v2 "匿名跳过" 语义一致。
	RoleAnonymous = ""
)

// minSecretLen 是 NewAuth 接受的最短密钥字节数。HS256 不应使用弱/可猜密钥;
// 此长度迫使调用方传入一个足够随机的密钥,而非默认空串。
const minSecretLen = 32

// ErrEmptySecret 在 NewAuth 收到空或过短密钥时返回。
var ErrEmptySecret = errors.New("ginext: auth secret must be at least 32 bytes")

// TokenType 区分 access 与 refresh 令牌。鉴权中间件只接受 access 令牌。
//
// 为使"未声明类型"可区分(避免 v2 TokenTypeAccess=0 零值导致缺 type 字段
// 的旧 token 被误判为 access),零值 TokenTypeUnset 不是 access:鉴权中间件
// 显式要求 claims.Type == TokenTypeAccess,任何其它值(含未声明)都被拒。
type TokenType int

const (
	// TokenTypeUnset 是零值,表示令牌未声明类型。鉴权中间件不接受此类型。
	TokenTypeUnset TokenType = 0
	// TokenTypeAccess 用于接口鉴权。
	TokenTypeAccess TokenType = 1
	// TokenTypeRefresh 用于换取新的 access 令牌,不接受其直接访问受保护资源。
	TokenTypeRefresh TokenType = 2
)

// CookieTokenName 是 cookie 模式下存取令牌所用的 cookie 名。
const CookieTokenName = "token"

// ExpiresInMinutes 是 cookie 模式续期令牌的有效期(分钟)。
const ExpiresInMinutes = 15

var (
	// ErrInvalidToken 表示令牌解析失败或不可信。
	ErrInvalidToken = errors.New("invalid token")
	// ErrUnauthorized 表示鉴权失败(缺失令牌、角色不匹配等)。
	ErrUnauthorized = errors.New("unauthorized")
)

// Claims 是 JWT 的业务载荷。Roles 是一个显式的角色字符串切片,鉴权时做
// 集合包含判定,不再使用 bit-mask。
type Claims struct {
	UserID int64     `json:"user_id"`
	Roles  []string  `json:"roles"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

// Auth 用一个显式的密钥签发与校验令牌。它取代了 v2 的包级可变 jwtKey:
// 密钥在构造时固定,后续只读,避免了 InitJWT 与 Generate/Verify 之间的
// 数据竞争。生产环境必须用 [NewAuth] 注入强密钥。
type Auth struct {
	secret []byte
	now    func() time.Time
}

// NewAuth 以给定 secret 构造一个 Auth。secret 必须非空且不少于
// [minSecretLen] 字节,否则返回 [ErrEmptySecret] —— 空密钥会让任何人都
// 能伪造合法 JWT。now 默认指向 time.Now,仅用于测试注入固定时钟
// (也会被 [Auth.VerifyToken] 的 [jwt.WithTimeFunc] 使用,使校验走同一时钟)。
// 返回 *Auth 以使中间件方法可被调用方持有。
func NewAuth(secret string) (*Auth, error) {
	if len([]byte(secret)) < minSecretLen {
		return nil, ErrEmptySecret
	}
	return &Auth{
		secret: []byte(secret),
		now:    time.Now,
	}, nil
}

// GenerateToken 以 userid、roles、token 类型与过期时间签发一个 JWT。
func (a *Auth) GenerateToken(userid int64, roles []string, tokenType TokenType, expiredAt time.Time) (string, error) {
	claims := &Claims{
		UserID: userid,
		Roles:  roles,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secret)
}

// VerifyToken 校验 tokenString 并返回其 Claims。校验收紧为:
//   - 仅接受 HS256([jwt.WithValidMethods]),杜绝 none/RS 等算法混淆;
//   - 必须带 exp([jwt.WithExpirationRequired]),拒绝无过期时间的 token;
//   - 时钟走 [Auth.now]([jwt.WithTimeFunc]),与签发/续期同一时钟,便于测试注入。
//
// 任何解析失败都包装为 [ErrUnauthorized],以便调用方用
// [IsUnauthorizedError] 统一识别。
func (a *Auth) VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (any, error) { return a.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(a.now),
	)
	if err != nil {
		return nil, errors.Join(ErrUnauthorized, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RequireRoleHeader 返回一个 gin 中间件,从 Authorization 头(可选
// "Bearer " 前缀)取 JWT 并校验。仅当令牌是 access 类型、且其 roles 与
// 要求的 requiredRoles 有交集时放行;否则以 401 中止。
//
// 当 requiredRoles 含 [RoleAnonymous] 或为空时,中间件放行所有请求
// (匿名跳过),与 v2 语义一致。鉴权成功后,user_id 与 roles 被写入
// gin context。
func (a *Auth) RequireRoleHeader(requiredRoles ...string) gin.HandlerFunc {
	return a.requireRole(requiredRoles, a.tokenFromHeader, false)
}

// RequireRoleCookie 返回一个 gin 中间件,从 cookie([CookieTokenName])
// 取 JWT 并校验,规则同 [RequireRoleHeader]。鉴权成功后会用
// [Auth.GenerateToken] 续期令牌并写回 cookie(滚动刷新),沿用 v2 行为。
func (a *Auth) RequireRoleCookie(requiredRoles ...string) gin.HandlerFunc {
	return a.requireRole(requiredRoles, a.tokenFromCookie, true)
}

// IsAuthedHeader 是 [RequireRoleHeader] 的主动判定版:不写响应、不中止,
// 仅在通过时把 user_id/roles 写入 gin context 并返回 true。用于业务
// 代码需要在同一 handler 内按登录态分支时。
func (a *Auth) IsAuthedHeader(c *gin.Context, requiredRoles ...string) bool {
	return a.isAuthed(c, requiredRoles, a.tokenFromHeader, false)
}

// IsAuthedCookie 是 [RequireRoleCookie] 的主动判定版:含 cookie 续期。
func (a *Auth) IsAuthedCookie(c *gin.Context, requiredRoles ...string) bool {
	return a.isAuthed(c, requiredRoles, a.tokenFromCookie, true)
}

// tokenFromHeader 从 Authorization 头取令牌,去掉可选 "Bearer " 前缀。
func (a *Auth) tokenFromHeader(c *gin.Context) string {
	tokenString := c.GetHeader("Authorization")
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimSpace(tokenString[7:])
	}
	return tokenString
}

// tokenFromCookie 从 [CookieTokenName] cookie 取令牌。
func (a *Auth) tokenFromCookie(c *gin.Context) string {
	tokenString, _ := c.Cookie(CookieTokenName)
	return tokenString
}

// requireRole 是 RequireRoleHeader/Cookie 的共享中间件实现。tokenFn
// 决定令牌来源;refreshCookie 为 true 时 cookie 模式在鉴权成功后续期。
func (a *Auth) requireRole(requiredRoles []string, tokenFn func(*gin.Context) string, refreshCookie bool) gin.HandlerFunc {
	if allowsAnonymous(requiredRoles) {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		le := logger.FromContext(c.Request.Context())
		claims, ok := a.verifyAndCheck(c, tokenFn, requiredRoles, le)
		if !ok {
			ErrorResponse(c, errcode.Unauthorized(http.StatusUnauthorized, "unauthorized"))
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("roles", claims.Roles)
		if refreshCookie && !a.refreshCookie(c, claims, le) {
			ErrorResponse(c, errcode.InternalServer(http.StatusInternalServerError, "failed to generate token"))
			return
		}
		c.Next()
	}
}

// isAuthed 是 RequireRoleHeader/Cookie 主动判定的共享实现。
func (a *Auth) isAuthed(c *gin.Context, requiredRoles []string, tokenFn func(*gin.Context) string, refreshCookie bool) bool {
	if allowsAnonymous(requiredRoles) {
		return true
	}
	le := logger.FromContext(c.Request.Context())
	claims, ok := a.verifyAndCheck(c, tokenFn, requiredRoles, le)
	if !ok {
		return false
	}
	c.Set("user_id", claims.UserID)
	c.Set("roles", claims.Roles)
	if refreshCookie && !a.refreshCookie(c, claims, le) {
		return false
	}
	return true
}

// refreshCookie 用 claims 续签一个 access 令牌并写回 cookie。返回 false
// 表示续签失败(调用方应视为鉴权失败)。沿用 v2 的滚动刷新策略。
func (a *Auth) refreshCookie(c *gin.Context, claims *Claims, le logger.Entry) bool {
	expiredAt := a.now().Add(ExpiresInMinutes * time.Minute)
	newToken, err := a.GenerateToken(claims.UserID, claims.Roles, TokenTypeAccess, expiredAt)
	if err != nil {
		le.WithError(err).Warnf("failed to generate token")
		return false
	}
	c.SetCookie(CookieTokenName, newToken, ExpiresInMinutes*60, "/", "", false, true)
	return true
}

// verifyAndCheck 是单一的令牌校验路径:取令牌 → VerifyToken → 校验
// TokenType 为 access → 校验角色集合相交。它取代了 v2 的 bit-mask 判定。
func (a *Auth) verifyAndCheck(c *gin.Context, tokenFn func(*gin.Context) string, requiredRoles []string, le logger.Entry) (*Claims, bool) {
	tokenString := tokenFn(c)
	if tokenString == "" {
		le.Warnf("missing token")
		return nil, false
	}

	claims, err := a.VerifyToken(tokenString)
	if err != nil {
		le.WithError(err).Warnf("error parsing token")
		return nil, false
	}

	if claims.Type != TokenTypeAccess {
		le.Warnf("invalid token type")
		return nil, false
	}

	if !roleMatches(claims.Roles, requiredRoles) {
		le.Warnf("role not match")
		return nil, false
	}

	return claims, true
}

// allowsAnonymous 报告 requiredRoles 是否意味着匿名放行(含 RoleAnonymous 或为空)。
func allowsAnonymous(requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}
	for _, r := range requiredRoles {
		if r == RoleAnonymous {
			return true
		}
	}
	return false
}

// roleMatches 报告 claims 的角色集合与要求集合是否有交集。
// 空要求集合已被 allowsAnonymous 在更上层处理,此处假定 required 非空。
func roleMatches(claimsRoles, requiredRoles []string) bool {
	have := make(map[string]struct{}, len(claimsRoles))
	for _, r := range claimsRoles {
		have[r] = struct{}{}
	}
	for _, want := range requiredRoles {
		if _, ok := have[want]; ok {
			return true
		}
	}
	return false
}

// IsUnauthorizedError 报告 err 是否属于鉴权失败类(缺失/无效令牌)。
// 供调用方在统一错误处理里识别 401 类错误。
func IsUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInvalidToken)
}
