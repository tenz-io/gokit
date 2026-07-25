# ginext

Gin 扩展：请求绑定+校验（基于 annotation）、JWT 鉴权、RPC 拦截器、结构化错误响应。

## 功能特性

- `BindAndValidate` 依据结构体 `bind`/`validate` 标签，自动从 uri、query、header、form、multipart、JSON 等来源绑定参数并统一校验
- `Response`/`ErrorResponse` 输出统一的 `ResponseFrame{code,message,data}` 响应体，错误会自动映射 `errcode.Error` 的状态码；实现 `FileResponse` 接口可直接返回文件流
- JWT 鉴权：`GenerateToken`/`VerifyToken` 生成校验令牌，`Authenticate`/`AuthenticateRest`/`AuthenticateCookie` 按角色（`RoleType`）生成鉴权中间件，Cookie 模式支持自动续期
- `IsAuthenticated`/`IsAuthenticateRest`/`IsAuthenticateCookie` 可在业务代码里主动判断当前请求是否已通过鉴权
- `AllRpcInterceptor` 内置一条责任链：Tracer → Metrics → 流量记录 → 慢日志 → Panic 恢复，也可用 `NewTracerRpcInterceptor` 等单个构造函数自由组合
- `IsUnauthorizedError` 用于识别鉴权失败类错误，方便统一处理

## 快速开始

```go
import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/tenz-io/gokit/ginext/v2"
)

type CreateUserReq struct {
    Name string `bind:"form,name=name" validate:"required"`
}

func main() {
    r := gin.Default()

    r.POST("/api/user", func(c *gin.Context) {
        var req CreateUserReq
        if err := ginext.BindAndValidate(c, &req); err != nil {
            ginext.ErrorResponse(c, err)
            return
        }
        ginext.Response(c, gin.H{"name": req.Name})
    })

    // 仅 admin 角色可访问，基于 Authorization 头的 JWT 鉴权
    r.GET("/api/admin", ginext.Authenticate(ginext.RoleAdmin, ginext.AuthTypeRest), func(c *gin.Context) {
        ginext.Response(c, gin.H{"ok": true})
    })

    // 登录成功后签发令牌
    r.POST("/api/login", func(c *gin.Context) {
        token, err := ginext.GenerateToken(1001, ginext.RoleUser, ginext.TokenTypeAccess, time.Now().Add(time.Hour))
        if err != nil {
            ginext.ErrorResponse(c, err)
            return
        }
        ginext.Response(c, gin.H{"token": token})
    })

    r.Run(":8080")
}
```

## API 速查

| 函数/类型 | 说明 |
|---|---|
| `BindAndValidate(c, ptr)` | 按 `bind`/`validate` 标签从各来源绑定请求参数并统一校验 |
| `Validate(c, ptr)` | 仅对已填充的结构体执行 annotation 校验 |
| `Response(c, data)` | 输出统一成功响应；`data` 实现 `FileResponse` 时直接返回文件 |
| `ErrorResponse(c, err, data...)` | 输出统一错误响应，自动映射 `errcode.Error` 状态码 |
| `ResponseFrame` | 统一响应体结构 `{code, message, data}` |
| `FileResponse` | 实现 `GetFile() []byte` 即可让 `Response` 返回文件流 |
| `InitJWT(secretKey)` | 设置 JWT 签名密钥（默认使用内置密钥） |
| `GenerateToken(userid, role, tokenType, expiredAt)` | 生成 JWT |
| `VerifyToken(tokenString)` | 校验并解析 JWT，返回 `Claims` |
| `Authenticate(role, authType)` | 按角色+鉴权方式（Cookie/Header）生成鉴权中间件 |
| `AuthenticateRest(role)` | 基于 `Authorization` 头的 JWT 鉴权中间件 |
| `AuthenticateCookie(role)` | 基于 Cookie 的 JWT 鉴权中间件，鉴权成功后自动续期 |
| `IsAuthenticated(c, role, authType)` / `IsAuthenticateRest` / `IsAuthenticateCookie` | 主动判断当前请求是否已通过鉴权 |
| `IsUnauthorizedError(err)` | 判断错误是否为未鉴权/无效 token |
| `RoleType`/`AuthType`/`TokenType` 及其常量 | 角色（匿名/管理员/用户）、鉴权方式（Web/Rest）、Token 类型（access/refresh）枚举 |
| `RpcHandler`/`RpcInterceptor` | RPC 处理函数类型与拦截器接口，`Intercept` 包裹业务 handler |
| `AllRpcInterceptor` | 内置拦截器链：Tracer→Metrics→流量记录→慢日志→Panic 恢复→业务 handler |
| `NewTracerRpcInterceptor`/`NewMetricsRpcInterceptor`/`NewTrafficRpcInterceptor`/`NewSlogLogRpcInterceptor`/`NewPanicRecoverRpcInterceptor` | 单个拦截器构造函数，可按需组合顺序（见 `NewRpcInterceptors`） |

引入路径：`github.com/tenz-io/gokit/ginext/v2`
