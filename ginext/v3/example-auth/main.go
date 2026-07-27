package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tenz-io/gokit/ginext/v3"
	"github.com/tenz-io/gokit/logger/v3"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithFilePath("log"),
		logger.WithCaller(true),
		logger.WithCallerSkip(1),
	)
}

func main() {
	// 生产环境务必注入强密钥(>=32 字节),此处示例值仅供本地运行。
	auth, err := ginext.NewAuth("change-me-to-a-strong-secret-32bytes!")
	if err != nil {
		log.Fatalf("invalid auth secret: %v", err)
	}

	engine := gin.New()

	// 登录成功后签发 access 令牌。
	engine.POST("/api/login", func(c *gin.Context) {
		token, err := auth.GenerateToken(1001, []string{ginext.RoleUser}, ginext.TokenTypeAccess, time.Now().Add(time.Hour))
		if err != nil {
			ginext.ErrorResponse(c, err)
			return
		}
		ginext.Response(c, gin.H{"token": token})
	})

	// 仅 admin 角色可访问,基于 Authorization 头的 JWT 鉴权。
	engine.GET("/api/admin", auth.RequireRoleHeader(ginext.RoleAdmin), func(c *gin.Context) {
		userid, _ := c.Get("user_id")
		ginext.Response(c, gin.H{"user_id": userid, "ok": true})
	})

	// 普通用户可访问,基于 cookie 的 JWT 鉴权(成功后自动续期)。
	engine.GET("/api/me", auth.RequireRoleCookie(ginext.RoleUser), func(c *gin.Context) {
		userid, _ := c.Get("user_id")
		ginext.Response(c, gin.H{"user_id": userid})
	})

	// 主动判定鉴权态(不写响应、不中止)。
	engine.GET("/api/optional", func(c *gin.Context) {
		if auth.IsAuthedHeader(c, ginext.RoleUser) {
			ginext.Response(c, gin.H{"authenticated": true})
			return
		}
		ginext.Response(c, gin.H{"authenticated": false})
	})

	log.Println("server is running on :8080")
	if err := engine.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
