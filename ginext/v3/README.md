# ginext

Gin 扩展：请求绑定+校验（基于 annotation）、JWT 鉴权、统一响应、RPC 观测拦截器链。v3 是一次无兼容包袱的干净重写：删掉 `metadata` 子包与 `init()` 装配、把可变全局 JWT 密钥收进显式 `Auth` 结构、用显式 `[]string` 角色替代脆弱的 bit-mask、拦截器改用函数类型与可配置 `Chain`。

```go
import "github.com/tenz-io/gokit/ginext/v3"
```

## 模块介绍

ginext 在 [gin-gonic/gin] 之上提供四件让业务 handler 更薄的事：

- **请求绑定与校验**（`BindAndValidate` / `Validate`）：依据结构体 `bind`/`validate`/`default` 标签，从 uri、query、header、form、multipart、JSON 自动绑定参数并统一校验，失败全部转 400。
- **统一响应**（`Response` / `ResponseStatus` / `ErrorResponse`）：输出 `{code,message,data}`；data 实现 `FileResponse` 时直接返回文件流；`ResponseStatus` 可表达 201/204 等非 200 成功状态；`ErrorResponse` 把 `errcode.Error` 自动映射到 HTTP 状态码。
- **JWT 鉴权**（`Auth` / `RequireRoleHeader` / `RequireRoleCookie`）：用显式密钥签发与校验令牌，基于显式角色集合做中间件鉴权，cookie 模式滚动续期。
- **RPC 观测拦截器链**（`Chain` / `DefaultChain` / `ChainContext`）：串联 PanicRecover→Tracer→Metrics→Traffic→SlowLog，用于非 HTTP 的 RPC 处理场景，可按需裁剪。

## 能力清单

| 能力 | 含义 |
|---|---|
| 多来源请求绑定 | `BindAndValidate` 依据 `bind` 标签自动从 uri、query、header、form、multipart、JSON 中取值填充结构体，无需手写各来源的解析代码 |
| 绑定字段默认值填充 | 绑定前先按 `default` 标签为字段填充默认值，用于可选参数（如分页 `offset`）缺省场景 |
| 校验失败统一转 400 错误 | 绑定/校验失败时自动收集全部 annotation 校验错误并包装成 `errcode.Error`（400），便于前端展示具体字段错误 |
| 多文件表单上传绑定 | multipart 请求下按 `bind:"file"` 标签读取上传文件内容并写入 `[]byte` 字段，同时支持表单文本字段混合绑定 |
| 统一成功/错误响应体 | `Response`/`ResponseStatus`/`ErrorResponse` 统一输出 `{code,message,data}` 结构，错误自动映射 `errcode.Error` 的 HTTP 状态码 |
| 可定制 HTTP 成功状态 | `ResponseStatus(c, status, data)` 在保持统一响应体的同时表达 201/204 等非 200 成功状态 |
| 文件流直传响应 | 响应数据实现 `FileResponse.GetFile()` 时，`Response`/`ResponseStatus` 会自动探测内容类型并直接返回文件二进制流 |
| JWT 令牌签发与校验 | `Auth.GenerateToken`/`Auth.VerifyToken` 基于 user ID、roles、token 类型（access/refresh）签发和解析 JWT，密钥在构造时固定 |
| 基于显式角色的鉴权中间件 | `RequireRoleHeader`/`RequireRoleCookie` 按显式角色集合做集合包含判定拦截请求，cookie 模式下鉴权成功自动续期 Token |
| 主动鉴权态判定 | `IsAuthedHeader`/`IsAuthedCookie` 在业务代码里主动判断当前请求是否已通过鉴权，不写响应、不中止 |
| RPC 拦截器链（含慢日志与 Panic 恢复） | `DefaultChain` 串联 Panic 恢复、Tracer 注入、Metrics 打点、流量记录、慢请求告警五个环节，用于非 HTTP 的 RPC 处理场景 |
| 可配置拦截器链 | `NewChain`/`Chain.With` 用函数式 `Interceptor` 自由组合与裁剪链顺序 |

## 快速开始

```go
import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/tenz-io/gokit/ginext/v3"
)

type CreateUserReq struct {
    Name string `bind:"form,name=name" validate:"required"`
}

func main() {
    r := gin.Default()
    auth := ginext.NewAuth("change-me-to-a-strong-secret")

    r.POST("/api/user", func(c *gin.Context) {
        var req CreateUserReq
        if err := ginext.BindAndValidate(c, &req); err != nil {
            ginext.ErrorResponse(c, err)
            return
        }
        ginext.ResponseStatus(c, http.StatusCreated, gin.H{"name": req.Name})
    })

    // 仅 admin 角色可访问,基于 Authorization 头的 JWT 鉴权
    r.GET("/api/admin", auth.RequireRoleHeader(ginext.RoleAdmin), func(c *gin.Context) {
        ginext.Response(c, gin.H{"ok": true})
    })

    // 登录成功后签发令牌
    r.POST("/api/login", func(c *gin.Context) {
        token, err := auth.GenerateToken(1001, []string{ginext.RoleUser}, ginext.TokenTypeAccess, time.Now().Add(time.Hour))
        if err != nil {
            ginext.ErrorResponse(c, err)
            return
        }
        ginext.Response(c, gin.H{"token": token})
    })

    // RPC 拦截器链:PanicRecover→Tracer→Metrics→Traffic→SlowLog
    r.POST("/api/rpc", func(c *gin.Context) {
        ctx := ginext.ChainContext(c, "rpc")
        resp, err := ginext.DefaultChain().Intercept(ctx, nil, func(ctx context.Context, req any) (any, error) {
            return gin.H{"hello": "world"}, nil
        })
        if err != nil {
            ginext.ErrorResponse(c, err)
            return
        }
        ginext.Response(c, resp)
    })

    r.Run(":8080")
}
```

## API 速查

| 函数/类型 | 说明 |
|---|---|
| `BindAndValidate(c, ptr)` | 按 `bind`/`validate` 标签从各来源绑定请求参数并统一校验 |
| `Validate(c, ptr)` | 仅对已填充的结构体执行 annotation 校验 |
| `Response(c, data)` | 以 HTTP 200 输出统一成功响应；`data` 实现 `FileResponse` 时直接返回文件 |
| `ResponseStatus(c, status, data)` | 以给定 HTTP status 输出统一成功响应（201/204 等）；文件流仍固定 200 |
| `ErrorResponse(c, err, data...)` | 输出统一错误响应，自动映射 `errcode.Error` 状态码 |
| `ResponseFrame` | 统一响应体结构 `{code, message, data}` |
| `FileResponse` | 实现 `GetFile() []byte` 即可让 `Response` 返回文件流 |
| `NewAuth(secret)` | 用显式密钥(>=32 字节)构造 `Auth`,返回 `(*Auth, error)`(空/过短返回 `ErrEmptySecret`) |
| `SetMaxBodyBytes(n)` | 覆盖 `BindAndValidate` 读取请求体的字节上限(默认 10 MiB);0/负值恢复默认 |
| `Auth.GenerateToken(userid, roles, tokenType, expiredAt)` | 生成 JWT |
| `Auth.VerifyToken(tokenString)` | 校验并解析 JWT(仅 HS256、要求 exp、走 `Auth.now`),返回 `Claims` |
| `Auth.RequireRoleHeader(roles...)` | 基于 `Authorization` 头的 JWT 鉴权中间件 |
| `Auth.RequireRoleCookie(roles...)` | 基于 Cookie 的 JWT 鉴权中间件，鉴权成功后自动续期 |
| `Auth.IsAuthedHeader(c, roles...)` / `Auth.IsAuthedCookie` | 主动判定当前请求是否已通过鉴权 |
| `IsUnauthorizedError(err)` | 判断错误是否为未鉴权/无效 token |
| `RoleAdmin`/`RoleUser`/`RoleAnonymous` | 显式角色字符串常量 |
| `TokenType` 及其常量 | token 类型（`TokenTypeAccess`/`TokenTypeRefresh`） |
| `Claims` | JWT 载荷：`UserID`、`Roles []string`、`Type` |
| `RpcHandler` / `Interceptor` | RPC 处理函数类型与拦截器函数类型 |
| `NewChain(interceptors...)` / `DefaultChain()` / `Chain.With(...)` | 构造/裁剪拦截器链 |
| `Chain.Intercept(ctx, req, handler)` | 执行链 |
| `ChainContext(c, cmd)` | 把 gin 请求边界信息（request ID/flag/logger/monitor）折叠进 context |
| `PanicRecoverInterceptor`/`TracerInterceptor`/`MetricsInterceptor`/`TrafficInterceptor`/`SlowLogInterceptor` | 单个拦截器构造函数,可按需组合顺序 |

## 测试

```bash
# 在仓库根(workspace 模式)
go test ./ginext/v3/...

# 模块独立模式
cd ginext/v3 && GOWORK=off go test ./... -cover
```

覆盖范围(模块独立模式):

| 包 | 覆盖率 |
|---|---|
| `ginext/v3` | ~88% |
| `ginext/v3/errcode` | 100% |

测试文件按职责拆分:

| 文件 | 覆盖范围 |
|---|---|
| `bind_test.go` | 文件上传 / form / json / multipart 绑定的成功路径 |
| `bind_more_test.go` | `Validate`、`warpError` 三路、校验失败→400、畸形 JSON→400、缺失必填 uri→400、非指针入参报错 |
| `bind_edge_test.go` | 缺失必填 query→400、uri/query 类型转换失败→400、form/multipart 方法不允许→400、multipart 无文件字段→400、空文件不报错、`default` 填充与不覆盖、Content-Type/方法跳过分支 |
| `bind_overlay_test.go` | JSON 不覆盖 uri/query/header、可选 form 字段类型错误→400、可选 JSON 字段类型错误→400、请求体超限→413 |
| `response_test.go` | `Response`(nil/normal/file)、`ResponseStatus`(201、文件流固定 200)、`ErrorResponse`(errcode 映射 / 普通 error→500 且不外泄) |
| `auth_test.go` | `RequireRoleHeader`/`RequireRoleCookie`/`IsAuthedHeader`、token 缺失/无效/有效、refresh 被拒、角色不匹配、cookie 续期 |
| `auth_more_test.go` | `VerifyToken` 成功/畸形/异密钥/无 exp 拒、`tokenFromHeader` Bearer 剥离、`IsAuthedCookie`、`roleMatches`、`allowsAnonymous`、不同密钥不同签名、`NewAuth` 拒弱密钥、unset type 被拒 |
| `handler_test.go` | `DefaultChain` 端到端、panic 恢复→500、空链、`Chain.With` 顺序 |
| `handler_more_test.go` | `Chain.With` 顺序、`getErrCode` 三态、`TracerInterceptor` 注入 request ID、`TrafficInterceptor` nil-safe、`SlowLogInterceptor` 快路径、`PanicRecoverInterceptor` 翻译 errcode、`cmdFromCtx` nil 安全 |
| `handler_panic_test.go` | panic 被 metrics 记为失败(code 非 0)、成功被记为 code=0(对照) |
| `errcode/error_test.go` | `Error.Error`、`New`(默认/显式 status)、9 个便捷构造函数、`FromError`(包装/非 errcode/nil) |

## 注意事项

- **`default` 不覆盖 `required`**:一个同时标了 `default:"1"` 与 `validate:"required"` 的字段,当绑定来源(uri/query/header)取不到值时,仍会因 `required` 报 "is required"。`default` 只对**非必填**的可选字段在缺失时填充。可选字段应写成 `bind:"query,name=offset" default:"0" validate:"gte=0"`(无 `required`)。
- **成功响应的 HTTP status**:`Response` 固定 200;需 201/204 等用 `ResponseStatus(c, status, data)`。文件流(`FileResponse`)无论传何 status,HTTP 均固定 200。
- **绑定/校验失败的 HTTP 状态**:经 `warpError` 包装为 `errcode.BadRequest`,HTTP status 与响应体 `code` 均为 400(v3 修复了 v2 status 仍为 200 的缺陷)。请求体超限返回 413。
- **JSON 不覆盖显式来源**:URI/Query/Header 字段一旦绑定即具权威性,JSON body 同名字段不得改写它们(防止 `/user/1` + `{"id":2}` 的 IDOR)。form/file 来源与 JSON 互斥(Content-Type 决定),不参与覆盖保护。
- **可选字段的解析错误必须 400**:可选(非 required)的 form/file 字段,只要值出现且解析/类型失败就返回 400,不再因非必填而静默置零。
- **请求体大小上限**:`BindAndValidate` 用 `http.MaxBytesReader` 限制请求体(默认 10 MiB),超限返回 413。用 `SetMaxBodyBytes(n)` 覆盖;大文件上传场景应在进入绑定器前自行处理或放宽上限。
- **内部错误不外泄**:`ErrorResponse` 对非 `errcode.Error` 的错误只回固定 `"internal server error"`,原始 `err` 经 `c.Error` 进入 gin 错误链供日志记录,绝不回显给客户端。
- **`Auth` 密钥**:`NewAuth(secret)` 拒绝空/过短(<32 字节)密钥并返回 `ErrEmptySecret`;密钥在构造后只读,生产环境务必注入强密钥。
- **JWT 校验收紧**:`VerifyToken` 仅接受 HS256(`jwt.WithValidMethods`)、要求 exp 存在(`WithExpirationRequired`)、时钟走 `Auth.now`(`WithTimeFunc`,便于测试注入);`TokenTypeAccess` 为非零值,缺 type 字段的旧 token 会被鉴权中间件拒。
- **panic 记为失败**:`DefaultChain` 把 `PanicRecoverInterceptor` 放在最内层(紧贴 handler),panic 先被转成 500 errcode 再向外展开,使 metrics/traffic/slowlog 的 defer 观察到非 nil err,从而把这次调用记为失败(而非误记成功)。

## 变更说明（v2 → v3）

- **删除 `metadata` 子包**：`metadata.MD` 持有 `*gin.Context` 造成耦合，且其 `c.Request.WithContext(...)` 未生效的 bug 一并消除。请求边界信息由 `ChainContext` 折叠进 `context.Context`，拦截器直接从 ctx 取，不再有中间 `MD` 层。
- **拦截器链改 `Chain`/`Interceptor` 函数类型**：`AllRpcInterceptor` 全局变量与 `init()` 倒序装配被 `DefaultChain()` + `NewChain(...)` 取代，可配、可测、可裁剪。
- **`Auth` 结构体替代全局 `jwtKey`**：密钥在构造时固定只读，消除 `InitJWT` 与 `Generate/Verify` 间的数据竞争；生产必须 `NewAuth` 注入强密钥。
- **显式 `[]string` 角色替代 bit-mask**：`RoleType=int32` + `RoleAnonymous=0` 的魔法值与 `role & claims.Role` 掩码判定被 `Claims.Roles []string` + 集合包含判定取代。
- **删 `AuthType` 枚举**：`Authenticate(role, authType)` 调度层被 `RequireRoleHeader`/`RequireRoleCookie` 两个直接函数取代。
- **`ResponseStatus` 支持 201/204**：v2 `Response` 恒 200，v3 新增 `ResponseStatus(c, status, data)`。
- **`tracer.ParseFlag` 替代 `getTracerFlag`**：复用 tracer/v3 的统一解析器，不再手写 `switch`。
- **修复 `warpError` HTTP 状态码**：v2 的 `warpError` 调 `errcode.New(http.StatusBadRequest, msg)` 把 400 当作业务 code，而 HTTP status 仍取默认 200，导致绑定/校验失败时 HTTP 返回 200。v3 改用 `errcode.BadRequest(...)`，使 code 与 status 都为 400。
- **安全加固（基于 code review）**：
  - `NewAuth` 改返回 `(*Auth, error)`，拒绝空/过短(<32 字节)密钥（v2 空密钥可被任意伪造 JWT）。
  - `VerifyToken` 加 `WithValidMethods("HS256")`/`WithExpirationRequired`/`WithTimeFunc`；`TokenType` 零值改为 `TokenTypeUnset`，`TokenTypeAccess` 为非零，杜绝缺 type 字段的旧 token 被当 access。
  - `BindAndValidate` 防止 JSON body 覆盖已绑定的 URI/Query/Header 字段；可选 form/file 字段解析错误不再被吞；`MaxBytesReader` 限制请求体(默认 10 MiB)防 OOM。
  - `ErrorResponse` 非 errcode 分支只回固定 `"internal server error"`，原始错误不外泄。
  - `DefaultChain` 链顺序调整(`PanicRecover` 最内层)，使 panic 被 metrics/traffic 记为失败。
- **`errcode` API 不变**：仅注释中文化，保持 `Response`/`ErrorResponse`/`warpError` 依赖。

## 发布检查清单

`go.mod` 中的 `replace` 指令(指向 `../../annotation/v3` 等 workspace 同级目录)与零时间伪版本 `v3.0.0-00010101000000-...` 是**未发布模块的临时解析方式**。下游 `go get` 不会继承本模块的 `replace`,因此**发布前必须**:

1. 先发布四个依赖模块并打 tag:`annotation/v3`、`logger/v3`、`monitor/v3`、`tracer/v3`。
2. 删除 `ginext/v3/go.mod` 末尾的 `replace (...)` 块。
3. 把 `require` 里的四个 `v3.0.0-00010101000000-000000000000` 改成真实版本(如 `v3.0.0`)。
4. 删除 `ginext/v3/go.sum`,运行 `go mod tidy` 重建。
5. 在 `go.work` 中移除本模块(发布后由 `go get` 解析,不再需要 workspace `use`)。
6. 同样处理 `example/`、`example-auth/` 两个示例模块的 `go.mod`(它们的 `replace` 也指向 `../../../` 同级目录)。

本模块的 `replace` 块已加注释说明"can be dropped once the modules are tagged"。在依赖发布前,workspace 模式下一切正常,无需改动。

引入路径：`github.com/tenz-io/gokit/ginext/v3`
