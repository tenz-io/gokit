// Package ginext 在 [github.com/gin-gonic/gin] 之上提供四件让业务 handler 更薄的事:
//
//   - 请求绑定与校验:[BindAndValidate] 依据结构体 `bind`/`validate`/`default`
//     标签,从 uri、query、header、form、multipart、JSON 自动绑定参数并统一校验。
//   - 统一响应:[Response]/[ResponseStatus] 输出 `{code,message,data}`;data 实现
//     [FileResponse] 时直接返回文件流;[ErrorResponse] 把 [errcode.Error] 自动映射到 HTTP 状态码。
//   - JWT 鉴权:[Auth] 以一个显式密钥签发与校验令牌,[RequireRoleHeader]/
//     [RequireRoleCookie] 提供基于显式角色集合的鉴权中间件。
//   - RPC 观测拦截器链:[Chain] 与 [DefaultChain] 串联 Tracer、Metrics、Traffic、
//     SlowLog、PanicRecover,用于非 HTTP 的 RPC 处理场景。
//
// # 快速开始
//
//	import (
//	    "time"
//
//	    "github.com/gin-gonic/gin"
//	    "github.com/tenz-io/gokit/ginext/v3"
//	)
//
//	type CreateUserReq struct {
//	    Name string `bind:"form,name=name" validate:"required"`
//	}
//
//	func main() {
//	    r := gin.Default()
//	    auth, err := ginext.NewAuth("change-me-to-a-strong-32-byte-secret!")
//	    if err != nil { log.Fatal(err) }
//
//	    r.POST("/api/user", func(c *gin.Context) {
//	        var req CreateUserReq
//	        if err := ginext.BindAndValidate(c, &req); err != nil {
//	            ginext.ErrorResponse(c, err)
//	            return
//	        }
//	        ginext.ResponseStatus(c, http.StatusCreated, gin.H{"name": req.Name})
//	    })
//
//	    r.GET("/api/admin", auth.RequireRoleHeader(ginext.RoleAdmin), func(c *gin.Context) {
//	        ginext.Response(c, gin.H{"ok": true})
//	    })
//
//	    r.POST("/api/login", func(c *gin.Context) {
//	        token, err := auth.GenerateToken(1001, []string{ginext.RoleUser}, ginext.TokenTypeAccess, time.Now().Add(time.Hour))
//	        if err != nil {
//	            ginext.ErrorResponse(c, err)
//	            return
//	        }
//	        ginext.Response(c, gin.H{"token": token})
//	    })
//
//	    r.Run(":8080")
//	}
//
// 引入路径:github.com/tenz-io/gokit/ginext/v3
package ginext
